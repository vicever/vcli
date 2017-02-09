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
	"errors"
	"fmt"
	"strings"
)

type overhead struct {
	grains     uint64
	header     offsets
	descriptor offsets
	rgd        offsets
	rgt        offsets
	gd         offsets
	gt         offsets
}

type offsets struct {
	first  uint64
	last   uint64
	length uint64
}

type Header struct {
	MagicNumber        uint32
	Version            uint32
	Flags              uint32
	Capacity           uint64
	GrainSize          uint64
	DescriptorOffset   uint64
	DescriptorSize     uint64
	NumGTEsPerGT       uint32
	RGDOffset          uint64
	GDOffset           uint64
	OverHead           uint64
	UncleanShutdown    byte
	SingleEndLineChar  byte
	NonEndLineChar     byte
	DoubleEndLineChar1 byte
	DoubleEndLineChar2 byte
	CompressAlgorithm  uint16
	Pad                [433]uint8
}

type descriptor struct {
}

type grainDirectory struct {
}

type grainTable struct {
}

func (build *builder) calculateOverhead() error {

	// grains within the grain table

	bytes := uint64(build.config.Disk.DiskSize * megabyte)
	sectors := ceiling(bytes, SectorSize)
	grains := ceiling(sectors, sectorsPerGrain)

	// calculate vmdk overhead size
	tables := ceiling(grains, tableMaxRows)
	dirSectors := ceiling(tables*ref32, SectorSize)

	// combine tables and directories, and double for redundancy
	tableAndDirSectors := 2 * (tables*tableSectors + dirSectors)

	// add vmdk descriptor size, then round up to nearest grain
	overhead := ceiling(tableAndDirSectors+headerSectors+descriptorSectors,
		sectorsPerGrain)

	build.header.Capacity = grains * sectorsPerGrain
	build.overhead.grains = overhead

	// header sectors
	build.overhead.header.first = 0
	build.overhead.header.length = 1
	build.overhead.header.last = build.overhead.header.first +
		build.overhead.header.length - 1

	// descriptor sectors
	build.overhead.descriptor.first = build.overhead.header.last + 1
	build.overhead.descriptor.length = 20
	build.overhead.descriptor.last = build.overhead.descriptor.first +
		build.overhead.descriptor.length - 1

	// redundant grain directory
	build.overhead.rgd.first = build.overhead.descriptor.last + 1
	build.overhead.rgd.length = dirSectors
	build.overhead.rgd.last = build.overhead.rgd.first +
		build.overhead.rgd.length - 1

	// redundant grain tables
	build.overhead.rgt.first = build.overhead.rgd.last + 1
	build.overhead.rgt.length = tables * tableSectors
	build.overhead.rgt.last = build.overhead.rgt.first +
		build.overhead.rgt.length - 1

	// grain directory
	build.overhead.gd.first = build.overhead.rgt.last + 1
	build.overhead.gd.length = dirSectors
	build.overhead.gd.last = build.overhead.gd.first +
		build.overhead.gd.length - 1

	// grain tables
	build.overhead.gt.first = build.overhead.gd.last + 1
	build.overhead.gt.length = tables * tableSectors
	build.overhead.gt.last = build.overhead.gt.first +
		build.overhead.gt.length - 1

	return nil

}

func (build *builder) populateSparseHeader() {

	build.header.MagicNumber = 0x564d444b
	build.header.Version = 1
	build.header.Flags = 0x3
	build.header.GrainSize = sectorsPerGrain
	build.header.DescriptorOffset = build.overhead.descriptor.first
	build.header.DescriptorSize = build.overhead.descriptor.length
	build.header.NumGTEsPerGT = tableMaxRows
	build.header.RGDOffset = build.overhead.rgd.first
	build.header.GDOffset = build.overhead.gd.first
	build.header.OverHead = build.overhead.grains * sectorsPerGrain
	build.header.SingleEndLineChar = '\n'
	build.header.NonEndLineChar = ' '
	build.header.DoubleEndLineChar1 = '\r'
	build.header.DoubleEndLineChar2 = '\n'
	build.header.CompressAlgorithm = 0

}

func (build *builder) populateStreamHeader() {

	build.header.MagicNumber = 0x564d444b
	build.header.Version = 3
	build.header.Flags = 0x30001
	build.header.GrainSize = sectorsPerGrain
	build.header.DescriptorOffset = build.overhead.descriptor.first
	build.header.DescriptorSize = build.overhead.descriptor.length
	build.header.NumGTEsPerGT = tableMaxRows
	build.header.RGDOffset = 0 // build.overhead.rgd.first
	build.header.GDOffset = build.overhead.rgd.first
	build.header.OverHead = build.overhead.grains * sectorsPerGrain
	build.header.SingleEndLineChar = '\n'
	build.header.NonEndLineChar = ' '
	build.header.DoubleEndLineChar1 = '\r'
	build.header.DoubleEndLineChar2 = '\n'
	build.header.CompressAlgorithm = 1

}

func (build *builder) generateSparseDescriptor() error {

	var err error
	firstGrain := build.overhead.descriptor.first / sectorsPerGrain
	lastGrain := build.overhead.descriptor.last / sectorsPerGrain

	if firstGrain != lastGrain {
		return errors.New("can't deal with oversized header")
	}

	buf := new(bytes.Buffer) //data[firstGrain].sectors[firstSector].data[:])

	uid := generateDiskUID()

	bytes := uint64(build.config.Disk.DiskSize * megabyte)
	sectors := ceiling(bytes, SectorSize)

	// function for writing to the vmdk buffer
	write := func(s string) {

		if err == nil {
			err = binary.Write(buf, binary.LittleEndian, []byte(s))
			if err != nil {
				err = errors.New("error writing header: %v" +
					err.Error())
			}
		}

	}

	write("# Disk DescriptorFile\n")
	write("version=1\n")
	write(fmt.Sprintf("CID=%s\n", strings.ToUpper(uid)))
	write("parentCID=ffffffff\n")
	write("createType=\"monolithicSparse\"\n\n")
	write("# Extent description\n")
	write(fmt.Sprintf("RW %d SPARSE \"%s\"\n\n", sectors,
		build.config.Name+".vmdk"))
	write("# The Disk Data Base\n")
	write("#DDB\n\n")
	write("ddb.virtualHWVersion = \"8\"\n")
	write("ddb.adapterType = \"ide\"\n")

	build.descriptor = buf

	return nil

}

func (build *builder) generateStreamDescriptor() error {

	var err error
	firstGrain := build.overhead.descriptor.first / sectorsPerGrain
	lastGrain := build.overhead.descriptor.last / sectorsPerGrain

	if firstGrain != lastGrain {
		return errors.New("can't deal with oversized header")
	}

	buf := new(bytes.Buffer) //data[firstGrain].sectors[firstSector].data[:])

	uid := generateDiskUID()

	bytes := uint64(build.config.Disk.DiskSize * megabyte)
	sectors := ceiling(bytes, SectorSize)

	// function for writing to the vmdk buffer
	write := func(s string) {

		if err == nil {
			err = binary.Write(buf, binary.LittleEndian, []byte(s))
			if err != nil {
				err = errors.New("error writing header: %v" +
					err.Error())
			}
		}

	}

	write("# Disk DescriptorFile\n")
	write("version=1\n")
	write(fmt.Sprintf("CID=%s\n", strings.ToUpper(uid)))
	write("parentCID=ffffffff\n")
	write("isNativeSnapshot=\"no\"\n")
	write("createType=\"streamOptimized\"\n\n")
	write("# Extent description\n")
	write(fmt.Sprintf("RW %d SPARSE \"%s\"\n\n", sectors,
		build.config.Name+".vmdk"))
	write("# The Disk Data Base\n")
	write("#DDB\n\n")
	write("ddb.virtualHWVersion = \"8\"\n")
	write("ddb.adapterType = \"ide\"\n")

	build.descriptor = buf

	return nil

}

func (build *builder) writeOverhead() error {

	data := make([]byte, build.overhead.grains*sectorsPerGrain*SectorSize, build.overhead.grains*sectorsPerGrain*SectorSize)

	// write header to data
	err := build.writeHeader(data)
	if err != nil {
		return err
	}

	// write descriptor to data
	err = build.writeDescriptor(data)
	if err != nil {
		return err
	}

	// write grain tables and directories to data
	err = build.writeGrainTablesAndDirectories(data, true)
	if err != nil {
		return err
	}

	// write redundant grain tables and directories to data
	err = build.writeGrainTablesAndDirectories(data, false)
	if err != nil {
		return err
	}

	// burn to disk
	b := make([]byte, build.overhead.grains*sectorsPerGrain*SectorSize,
		build.overhead.grains*sectorsPerGrain*SectorSize)
	buf := new(bytes.Buffer)

	err = binary.Write(buf, binary.LittleEndian, data)
	if err != nil {
		return err
	}

	copy(b, buf.Bytes())

	_, err = build.disk.WriteAt(b, 0)
	if err != nil {
		return err
	}

	return nil

}

func (build *builder) writeHeader(data []byte) error {

	firstGrain := build.overhead.header.first / sectorsPerGrain
	lastGrain := build.overhead.header.last / sectorsPerGrain

	if firstGrain != lastGrain {
		return errors.New("can't deal with oversized header")
	}

	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.LittleEndian, build.header)
	if err != nil {
		return err
	}

	copy(data, buf.Bytes())

	return nil

}

func (build *builder) writeDescriptor(data []byte) error {

	copy(data[SectorSize:], build.descriptor.Bytes())

	return nil

}

func (build *builder) writeStreamDescriptor(data []byte) error {

	copy(data[SectorSize:], build.descriptor.Bytes())

	return nil

}

func (build *builder) writeGrainTablesAndDirectories(data []byte, redundant bool) error {

	var gd, gt offsets

	if redundant {
		gd = build.overhead.rgd
		gt = build.overhead.rgt
	} else {
		gd = build.overhead.gd
		gt = build.overhead.gt
	}

	var err error
	var finished bool

	tables := gt.length / 4

	for i := uint64(0); i < tables && !finished; i++ {

		// add table entry to directory
		loc := gd.first*SectorSize + (i * ref32)
		dirBuf := new(bytes.Buffer)

		err = binary.Write(dirBuf, binary.LittleEndian, gt.first+i*ref32)
		if err != nil {
			return err
		}

		copy(data[loc:], dirBuf.Bytes())

		// write a table
		// tblBuf := new(bytes.Buffer)
		//
		// for j := uint64(0); j < tableMaxRows && !finished; i++ {
		//
		// 	index := i*tableMaxRows + j
		//
		// 	// if uint64(len(build.grains)) == index {
		// 	// 	finished = true
		// 	// 	break
		// 	// }
		//
		// 	// err = binary.Write(tblBuf, binary.LittleEndian, build.grains[index])
		// 	// if err != nil {
		// 	// 	return err
		// 	// }
		//
		// }
		//
		// copy(data[gt.first*sectorSize+i*ref32:], dirBuf.Bytes())

	}

	return nil

}
