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
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

func (env *Environment) BuildDynamicVHD(args *BuildArgs) (*os.File, error) {

	build := env.newBuilder(args)

	err := build.vhd()
	if err != nil && build.disk != nil {
		os.Remove(build.disk.Name())
	}

	return build.disk, err

}

func (build *builder) vhd() error {

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

	if build.config.Disk.DiskSize%2 != 0 {
		return fmt.Errorf("vhd disk size must be a multiple of 2MB")
	}

	conectix := uint64(0x636F6E6563746978)
	cxsparse := uint64(0x6378737061727365)
	timestamp := time.Now().Unix() - 946684800 // 2000 offset

	// CHS crap
	var cylinders, heads, sectorsPerTrack int
	var cylinderTimesHeads int

	totalSectors := build.config.Disk.DiskSize * 1024 * 2
	if totalSectors > 65535*16*255 {
		totalSectors = 65535 * 16 * 255
	}

	if totalSectors >= 65525*16*63 {
		sectorsPerTrack = 255
		heads = 16
		cylinderTimesHeads = totalSectors / sectorsPerTrack
	} else {
		sectorsPerTrack = 17
		cylinderTimesHeads = totalSectors / sectorsPerTrack
		heads = (cylinderTimesHeads + 1023) / 1024
		if heads < 4 {
			heads = 4
		}
		if cylinderTimesHeads >= (heads*1024) || heads > 16 {
			sectorsPerTrack = 31
			heads = 16
			cylinderTimesHeads = totalSectors / sectorsPerTrack
		}
		if cylinderTimesHeads >= heads*1024 {
			sectorsPerTrack = 63
			heads = 16
			cylinderTimesHeads = totalSectors / sectorsPerTrack
		}
	}
	cylinders = cylinderTimesHeads / heads

	// copy of hard disk footer
	footer := &VHDFooter{
		Cookie:             conectix,
		Features:           0x00000002,
		FileFormatVersion:  0x00010000,
		DataOffset:         512,
		TimeStamp:          uint32(timestamp),
		CreatorApplication: 0x76636C69,
		CreatorVersion:     0x00010000, // TODO: does this matter?
		CreatorHostOS:      0x5769326B, // TODO: does this matter?
		OriginalSize:       uint64(build.config.Disk.DiskSize * 1024 * 1024),
		CurrentSize:        uint64(build.config.Disk.DiskSize * 1024 * 1024),
		DiskGeometry:       uint32(cylinders<<16 | heads<<8 | sectorsPerTrack),
		DiskType:           3,
		// TODO: UniqueID
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, footer)
	if err != nil {
		return err
	}

	var checksum uint32

	for i := 0; i < buf.Len(); i++ {
		var x byte
		x, err = buf.ReadByte()
		if err != nil {
			return err
		}
		checksum += uint32(x) // TODO: does this achieve one's complement?
	}

	footer.Checksum = checksum

	fbuf := new(bytes.Buffer)
	err = binary.Write(fbuf, binary.BigEndian, footer)
	if err != nil {
		return err
	}

	_, err = build.disk.WriteAt(fbuf.Bytes(), 0) // header goes at offset 512
	if err != nil {
		return err
	}

	// dynamic disk header
	header := &VHDHeader{
		Cookie:          cxsparse,
		DataOffset:      0xFFFFFFFFFFFFFFFF,
		TableOffset:     1536,
		HeaderVersion:   0x00010000,
		MaxTableEntries: uint32(build.config.Disk.DiskSize / 2), // TODO: check non odd number for disk size
		BlockSize:       0x200000,
	}

	buf = new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, header)
	if err != nil {
		return err
	}

	checksum = 0

	for i := 0; i < buf.Len(); i++ {
		var x byte
		x, err = buf.ReadByte()
		if err != nil {
			return err
		}
		checksum += uint32(x) // TODO: does this achieve one's complement?
	}

	header.Checksum = checksum

	buf = new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, header)
	if err != nil {
		return err
	}

	_, err = build.disk.WriteAt(buf.Bytes(), 512) // header goes at offset 512
	if err != nil {
		return err
	}

	// BAT (block allocation table)
	batEntries := build.config.Disk.DiskSize / 2
	batSize := ceiling(uint64(4*batEntries), 512) * 512

	dataStart := 1024 + 512 + batSize

	build.disk.WriteAt(bytes.Repeat([]byte{255}, int(dataStart*4)), int64(dataStart-batSize))

	// blocks

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
	err = build.writeVHDBlocks(dataStart)
	if err != nil {
		return fmt.Errorf("error writing grains to disk: %v", err)
	}

	// hard disk footer
	build.disk.Seek(0, 2)
	build.disk.Write(fbuf.Bytes())

	return nil

}

type VHDFooter struct { // 512 bytes
	Cookie             uint64
	Features           uint32
	FileFormatVersion  uint32
	DataOffset         uint64
	TimeStamp          uint32
	CreatorApplication uint32
	CreatorVersion     uint32
	CreatorHostOS      uint32
	OriginalSize       uint64
	CurrentSize        uint64
	DiskGeometry       uint32
	DiskType           uint32
	Checksum           uint32
	UniqueID           [16]byte
	SavedState         byte
	Reserved           [427]byte
}

type VHDHeader struct { // 1024 bytes
	Cookie              uint64
	DataOffset          uint64
	TableOffset         uint64
	HeaderVersion       uint32
	MaxTableEntries     uint32
	BlockSize           uint32
	Checksum            uint32
	ParentUniqueID      [16]byte
	ParentTimeStamp     uint32
	Reserved            [4]byte
	ParentUnicodeName   [512]byte
	ParentLocatorEntry1 [24]byte
	ParentLocatorEntry2 [24]byte
	ParentLocatorEntry3 [24]byte
	ParentLocatorEntry4 [24]byte
	ParentLocatorEntry5 [24]byte
	ParentLocatorEntry6 [24]byte
	ParentLocatorEntry7 [24]byte
	ParentLocatorEntry8 [24]byte
	Reserved2           [256]byte
}

func (build *builder) writeVHDBlocks(dataStart uint64) error {

	// re-open content file
	content, err := os.Open(build.content.file.Name())
	if err != nil {
		return err
	}

	i := 0
	for unfinished := true; unfinished; i++ {

		bsize := 2 * 1024 * 1024
		block := make([]byte, bsize, bsize)
		_, err = content.ReadAt(block, int64(i*bsize))

		if err != nil {

			if err != io.EOF {
				return err
			}

			// last content grain read
			unfinished = false

		}

		if !unfinished {
			copy(block[bsize-len(build.finalGrain):], build.finalGrain)
		}

		// write grain to disk
		err = build.writeVHDBlock(dataStart, i, block)
		if err != nil {
			return err
		}

	}

	// lastBlockNo := build.config.Disk.DiskSize / 2
	// err = build.writeVHDBlock(dataStart, lastBlockNo, build.finalGrain) // TODO: fix this for block instead of grain
	// if err != nil {
	// 	return err
	// }

	return nil

}

func (build *builder) writeVHDBlock(dataStart uint64, blockNo int, block []byte) error {

	empty := true
	for _, x := range block {
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

	// write grain to disk
	offset := int64(dataStart + uint64(entry*(2*1024*1024+512)))

	_, err := build.disk.WriteAt(bytes.Repeat([]byte{255}, 512), offset)
	if err != nil {
		return err
	}

	_, err = build.disk.WriteAt(block, offset+512)
	if err != nil {
		return err
	}

	// add entry to grain tables
	b := make([]byte, ref32)
	binary.BigEndian.PutUint32(b, uint32(offset/512))

	_, err = build.disk.WriteAt(b, int64(1024+512+blockNo*4))
	if err != nil {
		return err
	}

	return nil

}
