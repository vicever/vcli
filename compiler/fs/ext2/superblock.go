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
package ext2

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/sisatech/sherlock"
)

const (
	superUID = 1000
	superGID = 1000
)

type Superblock struct {
	TotalInodes         uint32
	TotalBlocks         uint32
	ReservedBlocks      uint32
	UnallocatedBlocks   uint32
	UnallocatedInodes   uint32
	SuperblockNumber    uint32
	BlockSize           uint32
	FragmentSize        uint32
	BlocksPerGroup      uint32
	FragmentsPerGroup   uint32
	InodesPerGroup      uint32
	LastMountTime       uint32
	LastWrittenTime     uint32
	MountsSinceCheck    uint16
	MountsCheckInterval uint16
	Signature           uint16
	State               uint16
	ErrorProtocol       uint16
	VersionMinor        uint16
	TimeLastCheck       uint32
	TimeCheckInterval   uint32
	OS                  uint32
	VersionMajor        uint32
	SuperUser           uint16
	SuperGroup          uint16
	Padding             [blockSize - 84]byte
}

func (super *Superblock) init(now time.Time) {

	super.SuperblockNumber = 1
	super.LastMountTime = uint32(now.Unix())
	super.LastWrittenTime = uint32(now.Unix())
	super.MountsCheckInterval = 20
	super.Signature = 0xEF53
	super.State = 1
	super.VersionMajor = 0
	super.VersionMinor = 0
	super.TimeLastCheck = uint32(now.Unix())
	super.SuperUser = superUID
	super.SuperGroup = superGID

}

func (ins *Instructions) writeSuperblock() {

	ins.superblock.UnallocatedInodes = ins.superblock.TotalInodes - ins.inodes

	// total
	//	- overhead
	//	- group_overhead * number_of_groups
	// 	- blocks used for data
	// 	- blocks lost to overlapping the end of the file
	ins.superblock.UnallocatedBlocks = ins.superblock.TotalBlocks -
		ins.overhead -
		ins.groupOverhead*ins.totalGroups -
		ins.blocks

	for i := uint32(0); i < ins.totalGroups; i++ {

		ins.superblock.SuperblockNumber = 1 + i*ins.blocksPerGroup

		buf := new(bytes.Buffer)

		sherlock.Check(binary.Write(buf, binary.LittleEndian, ins.superblock))

		ins.instructions = append(ins.instructions, &instruction{
			offset: int64((uint32(1) + i*ins.blocksPerGroup) * blockSize),
			length: blockSize,
			data:   buf,
		})

	}

}
