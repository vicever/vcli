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
package rawsparse

import (
	"fmt"
	"io"
	"os"
)

func (env *Environment) BuildSparse(args *BuildArgs) (*os.File, error) {

	build := env.newBuilder(args)

	err := build.sparse()
	if err != nil && build.disk != nil {
		os.Remove(build.disk.Name())
	}
	return build.disk, err

}

func (build *builder) sparse() error {

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

	build.log("Capacity: %v MB", build.config.Disk.DiskSize)

	// calculate required number of grains on the disk
	err = build.calculateOverhead()
	if err != nil {
		return fmt.Errorf("error analysing files: %v", err)
	}

	// initialize sparse header
	// build.populateSparseHeader()
	//
	// err = build.generateSparseDescriptor()
	// if err != nil {
	// 	return err
	// }

	// burn vmdk overhead to disk at the end
	// err = build.writeOverhead()
	// if err != nil {
	// 	return fmt.Errorf("error writing vmdk overhead: %v", err)
	// }

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
	err = build.writeSparseGrains()
	if err != nil {
		return fmt.Errorf("error writing grains to disk: %v", err)
	}

	return nil

}

func (build *builder) writeSparseGrains() error {

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
		err = build.writeGrain(i, grain)
		if err != nil {
			return err
		}

	}

	lastGrainNo := build.config.Disk.DiskSize*megabyte/SectorSize/sectorsPerGrain - 1
	err = build.writeLastGrain(lastGrainNo, build.finalGrain)
	if err != nil {
		return err
	}

	return nil

}

func (build *builder) writeLastGrain(grainNo int, grain []byte) error {

	entry := build.grainCounter - 1
	// write grain to disk
	offset := int64(SectorSize * sectorsPerGrain * (entry))
	_, err := build.disk.WriteAt(grain, offset)
	if err != nil {
		return err
	}

	return nil

}

func (build *builder) writeGrain(grainNo int, grain []byte) error {

	// empty := true
	// for _, x := range grain {
	// 	if x != 0 {
	// 		empty = false
	// 		break
	// 	}
	// }

	// if empty {
	// 	return nil
	// }

	entry := build.grainCounter
	build.grainCounter++
	// write grain to disk
	offset := int64(SectorSize * sectorsPerGrain * (entry))
	_, err := build.disk.WriteAt(grain, offset)
	if err != nil {
		return err
	}

	// // add entry to grain tables
	// b := make([]byte, ref32)
	// binary.LittleEndian.PutUint32(b, uint32(offset/SectorSize))
	//
	// _, err = build.disk.WriteAt(b, int64(build.overhead.gt.first*SectorSize+uint64(grainNo)*ref32))
	// if err != nil {
	// 	return err
	// }
	//
	// _, err = build.disk.WriteAt(b, int64(build.overhead.rgt.first*SectorSize+uint64(grainNo)*ref32))
	// if err != nil {
	// 	return err
	// }

	return nil

}
