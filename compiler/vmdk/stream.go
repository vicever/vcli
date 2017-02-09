// Copyright 2016 Sisa-Tech Pty Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package vmdk

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func (env *Environment) BuildStreamOptimized(args *BuildArgs) (*os.File, error) {

	build := env.newBuilder(args)

	err := build.stream()
	if err != nil && build.disk != nil {
		os.Remove(build.disk.Name())
	}

	return build.disk, err

}

func (build *builder) stream() error {

	// create new file to burn disk
	err := build.newDisk()
	if err != nil {
		return fmt.Errorf("error creating file for output: %v", err)
	}

	// validate args
	err = build.validateArgs()
	if err != nil {
		return fmt.Errorf("error validating arguments: %v", err)
	}

	// calculate required number of grains on the disk
	err = build.calculateOverhead()
	if err != nil {
		return fmt.Errorf("error analysing files: %v", err)
	}

	// initialize sparse header
	build.populateStreamHeader()

	err = build.generateStreamDescriptor()
	if err != nil {
		return err
	}

	// burn vmdk overhead to disk at the end
	err = build.writeOverhead()
	if err != nil {
		return fmt.Errorf("error writing vmdk overhead: %v", err)
	}

	// calculate total LBAs
	err = build.calculateLBAs()
	if err != nil {
		return fmt.Errorf("error analysing files: %v", err)
	}
	defer os.Remove(build.files.Name())

	// build disk contents
	err = build.diskContents()
	if err != nil {
		return fmt.Errorf("error compiling disk contents: %v", err)
	}
	defer os.Remove(build.content.file.Name())

	// write sparse grains
	build.seek = int64(build.overhead.grains * SectorSize * sectorsPerGrain)
	err = build.writeStreamOptimizedGrains()
	if err != nil {
		return fmt.Errorf("error writing grains to disk: %v", err)
	}

	return nil

}

func (build *builder) writeStreamOptimizedGrains() error {

	// re-open content file
	content, err := os.Open(build.content.file.Name())
	if err != nil {
		return err
	}

	i := 0
	for unfinished := true; unfinished; i++ {

		gsize := SectorSize * sectorsPerGrain
		grain := make([]byte, gsize, gsize)
		_, err = content.ReadAt(grain, int64(i*gsize))

		if err != nil {

			if err != io.EOF {
				return err
			}

			// last content grain read
			unfinished = false

		}

		// write grain to disk
		err = build.writeCompressedGrain(i, grain)
		if err != nil {
			return err
		}

	}

	lastGrainNo := build.config.Disk.DiskSize*megabyte/SectorSize/sectorsPerGrain - 1
	err = build.writeCompressedGrain(lastGrainNo, build.finalGrain)
	if err != nil {
		return err
	}

	err = build.writeFooter()
	if err != nil {
		return err
	}

	err = build.writeEOS()
	if err != nil {
		return err
	}

	return nil

}

type GrainMarker struct {
	LBA  uint64
	Size uint32
}

func (build *builder) writeCompressedGrain(grainNo int, grain []byte) error {

	// skip if grain is empty
	empty := true
	for _, x := range grain {
		if x != 0 {
			empty = false
			break
		}
	}

	if empty {
		return nil
	}

	entry := build.grainCounter
	build.grainCounter++

	// compress grain
	compressed, err := compress(grain)
	if err != nil {
		return err
	}

	// write grain marker
	offset := build.seek / SectorSize
	lba := int64(sectorsPerGrain * entry)

	marker := new(GrainMarker)
	marker.LBA = uint64(lba)
	marker.Size = uint32(len(compressed))

	b := make([]byte, 12)
	buf := new(bytes.Buffer)

	err = binary.Write(buf, binary.LittleEndian, marker)
	if err != nil {
		return err
	}

	copy(b, buf.Bytes())

	_, err = build.disk.WriteAt(b, build.seek)
	if err != nil {
		return err
	}
	build.seek = build.seek + int64(len(b))

	// write grain
	_, err = build.disk.WriteAt(compressed, build.seek)
	if err != nil {
		return err
	}
	build.seek = build.seek + int64(len(compressed))

	// pad to sector
	pad := SectorSize - (12+len(compressed))%SectorSize
	_, err = build.disk.WriteAt(make([]byte, pad, pad), build.seek)
	if err != nil {
		return err
	}
	build.seek = build.seek + int64(pad)

	// add entry to grain tables
	b = make([]byte, ref32)
	binary.LittleEndian.PutUint32(b, uint32(offset))

	_, err = build.disk.WriteAt(b, int64(build.overhead.gt.first*SectorSize+uint64(grainNo)*ref32))
	if err != nil {
		return err
	}

	_, err = build.disk.WriteAt(b, int64(build.overhead.rgt.first*SectorSize+uint64(grainNo)*ref32))
	if err != nil {
		return err
	}

	return nil

}

func compress(grain []byte) ([]byte, error) {

	buf := new(bytes.Buffer)

	// w, err := flate.NewWriter(buf, -1)
	// if err != nil {
	// 	return nil, err
	// }

	w := zlib.NewWriter(buf)

	_, err := w.Write(grain)
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

type MetadataMarker struct {
	Sectors uint64
	Size    uint32
	Type    uint32
	Pad     [496]byte
}

func (build *builder) writeFooter() error {

	// write footer marker
	marker := new(MetadataMarker)
	marker.Sectors = 0
	marker.Type = 3

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, marker)
	if err != nil {
		return err
	}

	_, err = build.disk.WriteAt(buf.Bytes(), build.seek)
	if err != nil {
		return err
	}
	build.seek = build.seek + int64(len(buf.Bytes()))

	// write footer

	// pad to sector
	// pad := sectorSize - (12+len(compressed))%sectorSize
	// _, err = build.disk.WriteAt(make([]byte, pad, pad), build.seek)
	// if err != nil {
	// 	return err
	// }
	// build.seek = build.seek + int64(pad)

	return nil

}

type EOSMarker struct {
	Val  uint64
	Size uint32
	Type uint32
	Pad  [496]byte
}

func (build *builder) writeEOS() error {

	marker := new(EOSMarker)

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, marker)
	if err != nil {
		return err
	}

	_, err = build.disk.WriteAt(buf.Bytes(), build.seek)
	if err != nil {
		return err
	}
	build.seek = build.seek + int64(len(buf.Bytes()))

	return nil

}
