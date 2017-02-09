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

func (ins *Instructions) writeBlockGroups() {

	for i := uint32(0); i < ins.totalGroups; i++ {

		ins.writeBlockGroup(i)

	}

}

func (ins *Instructions) writeBlockGroup(x uint32) {

	ins.writeBlockBitmap(x)
	ins.writeInodeBitmap(x)
	ins.writeInodeTable(x)

}

func (ins *Instructions) writeBlockBitmap(x uint32) {

	bitmap := make([]byte, blockSize)

	// overhead
	for i := uint32(0); i < ins.groupOverhead; i++ {

		bitmap[i/8] = bitmap[i/8] | 1<<(i%8)

	}

	// data
	var b uint32
	db := ins.blocksPerGroup - ins.groupOverhead
	y := int64(ins.blocks) - int64(x*db)
	if y > 0 {
		if uint32(y) >= db {
			b = db
		} else {
			b = uint32(y)
		}
	}

	for i := uint32(ins.groupOverhead); i < ins.groupOverhead+b; i++ {

		bitmap[i/8] = bitmap[i/8] | 1<<(i%8)

	}

	// overlap
	if x == ins.totalGroups-1 {

		// number of overlap blocks
		overlap := ins.overhead

		for i := uint32(ins.blocksPerGroup - overlap); i < ins.blocksPerGroup; i++ {

			bitmap[i/8] = bitmap[i/8] | 1<<(i%8)

		}

	}

	// unused bitmap region
	for i := uint32(ins.blocksPerGroup); i < blockSize*8; i++ {

		bitmap[i/8] = bitmap[i/8] | 1<<(i%8)

	}

	// write
	ins.instructions = append(ins.instructions, &instruction{
		offset: int64(blockSize * (ins.overhead + x*ins.blocksPerGroup + 1 + ins.blocksForBGDT)),
		length: blockSize,
		data:   bytes.NewBuffer(bitmap),
	})

}

func (ins *Instructions) writeInodeBitmap(x uint32) {

	bitmap := make([]byte, blockSize)

	// data
	var b uint32
	y := int64(ins.inodes) - int64(x*ins.inodesPerGroup)
	if y > 0 {
		if y >= int64(ins.inodesPerGroup) {
			b = ins.inodesPerGroup
		} else {
			b = uint32(y)
		}
	}

	for i := uint32(0); i < b; i++ {

		bitmap[i/8] = bitmap[i/8] | 1<<(i%8)

	}

	// unused bitmap region
	for i := uint32(ins.inodesPerGroup); i < blockSize*8; i++ {

		bitmap[i/8] = bitmap[i/8] | 1<<(i%8)

	}

	// write
	ins.instructions = append(ins.instructions, &instruction{
		offset: int64(blockSize * (ins.overhead + x*ins.blocksPerGroup + 1 + ins.blocksForBGDT + 1)),
		length: blockSize,
		data:   bytes.NewBuffer(bitmap),
	})

}

func (ins *Instructions) writeInodeTable(x uint32) {

	buf := new(bytes.Buffer)

	// data
	var b uint32
	y := int64(ins.inodes) - int64(x*ins.inodesPerGroup)
	if y > 0 {
		if y >= int64(ins.inodesPerGroup) {
			b = ins.inodesPerGroup
		} else {
			b = uint32(y)
		}
	}

	for i := uint32(0); i < b; i++ {

		sherlock.Check(binary.Write(buf, binary.LittleEndian, ins.nodes[i+x*ins.inodesPerGroup]))

	}

	for i := 0; i < buf.Len(); i = i + blockSize {

		var l int

		if i+blockSize > buf.Len() {
			l = buf.Len() % blockSize
		} else {
			l = blockSize
		}

		// write
		ins.instructions = append(ins.instructions, &instruction{
			offset: int64(blockSize*(ins.overhead+x*ins.blocksPerGroup+1+ins.blocksForBGDT+2) + uint32(i)),
			length: int64(l),
			data:   bytes.NewBuffer(buf.Bytes()[i : i+l]),
		})

	}

}
