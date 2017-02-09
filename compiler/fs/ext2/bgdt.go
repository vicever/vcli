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

	"github.com/sisatech/sherlock"
)

type BlockGroupDescriptor struct {
	BlockBitmap       uint32
	InodeBitmap       uint32
	InodeTable        uint32
	UnallocatedBlocks uint16
	UnallocatedInodes uint16
	Directories       uint16 // TODO
	Padding           [14]byte
}

func (ins *Instructions) writeBGDT() {

	bgdt := new(bytes.Buffer)

	for i := 0; uint32(i) < ins.totalGroups; i++ {

		groupOffset := ins.overhead + ins.blocksPerGroup*uint32(i)

		var spareBlocks uint16

		dataBlocksPerGroup := ins.blocksPerGroup - ins.groupOverhead

		// TODO double check this logic with blocks start at 0 instead of 1
		if ins.blocks > uint32(i+1)*dataBlocksPerGroup {
			spareBlocks = uint16(0)
		} else if ins.blocks > uint32(i)*dataBlocksPerGroup {
			spareBlocks = uint16(dataBlocksPerGroup - ins.blocks%dataBlocksPerGroup)
		} else {
			spareBlocks = uint16(dataBlocksPerGroup)
		}

		// overlap
		if uint32(i) == ins.totalGroups-1 {

			// number of overlap blocks
			overlap := ins.overhead

			spareBlocks -= uint16(overlap)

		}

		// spare inodes
		var spareInodes uint16

		// TODO double check this logic with inodes start at 1 instead of 0
		if ins.inodes > uint32(i+1)*ins.inodesPerGroup {
			spareInodes = uint16(0)
		} else if ins.inodes > uint32(i)*ins.inodesPerGroup {
			spareInodes = uint16(ins.inodesPerGroup - ins.inodes%ins.inodesPerGroup)
		} else {
			spareInodes = uint16(ins.inodesPerGroup)
		}

		// write
		buf := new(bytes.Buffer)

		sherlock.Check(binary.Write(buf, binary.LittleEndian, &BlockGroupDescriptor{
			BlockBitmap:       groupOffset + 1 + ins.blocksForBGDT,
			InodeBitmap:       groupOffset + 2 + ins.blocksForBGDT,
			InodeTable:        groupOffset + 3 + ins.blocksForBGDT,
			UnallocatedBlocks: spareBlocks,
			UnallocatedInodes: spareInodes,
			Directories:       uint16(ins.groupDirs[i]),
		}))

		sherlock.Check(bgdt.Write(buf.Bytes()))

	}

	for grp := uint32(0); grp < ins.totalGroups; grp++ {

		for i := 0; i < bgdt.Len(); i = i + blockSize {

			var l int

			if i+blockSize > bgdt.Len() {
				l = bgdt.Len() % blockSize
			} else {
				l = blockSize
			}

			// write
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(blockSize*(ins.overhead+grp*ins.blocksPerGroup+1) + uint32(i)),
				length: int64(l),
				data:   bytes.NewBuffer(bgdt.Bytes()[i : i+l]),
			})

		}

	}

}
