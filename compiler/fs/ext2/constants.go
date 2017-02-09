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

import "time"

const (
	blockSize = 1024

	maxBlocksPerGroup = 8192
	maxInodesPerGroup = 8192

	inodeEntrySize       = 128
	inodeEntriesPerBlock = blockSize / inodeEntrySize

	BGDTEntrySize       = 32
	BGDTEntriesPerBlock = blockSize / BGDTEntrySize
)

type constants struct {
	timestamp time.Time

	totalBlocks    uint32
	minInodes      uint32
	totalGroups    uint32
	inodesPerGroup uint32
	blocksPerGroup uint32
	blocksForBGDT  uint32

	overhead           uint32
	inodeTableOverhead uint32
	groupOverhead      uint32

	superblock Superblock
}

func (c *constants) compute(blocks, inodes uint32) {

	c.timestamp = time.Now()

	c.totalBlocks = blocks
	c.minInodes = inodes

	c.superblock.init(c.timestamp)

	c.computeNumberOfBlockGroups()
	c.computeInodesPerBlockGroup()
	c.computeBlocksPerBlockGroup()

	c.superblock.TotalBlocks = c.totalBlocks
	c.superblock.TotalInodes = c.inodesPerGroup * c.totalGroups
	c.superblock.BlocksPerGroup = c.blocksPerGroup
	c.superblock.InodesPerGroup = c.inodesPerGroup
	c.superblock.FragmentsPerGroup = c.blocksPerGroup
	c.superblock.UnallocatedBlocks = c.superblock.TotalBlocks
	c.superblock.UnallocatedInodes = c.superblock.TotalInodes
	c.superblock.ReservedBlocks = 0 // TODO?

}

func (c *constants) computeNumberOfBlockGroups() {

	// minimum number due to number of blocks
	min1 := ceiling(int64(c.totalBlocks), int64(maxBlocksPerGroup))

	// minimum number due to number of inodes
	min2 := ceiling(int64(c.minInodes), int64(maxInodesPerGroup))

	// take the higher of the two minimums
	min := min1
	if min2 > min {
		min = min2
	}

	c.totalGroups = uint32(min)

	c.blocksForBGDT = uint32(ceiling(int64(c.totalGroups), int64(BGDTEntriesPerBlock)))

	c.overhead = 1

}

func (c *constants) computeInodesPerBlockGroup() {

	// determine minimum inodes per bg
	min := ceiling(int64(c.minInodes), int64(c.totalGroups))

	// round up to nearest 8 to avoid wasting space
	min = align(min, inodeEntriesPerBlock)

	c.inodesPerGroup = uint32(min)

	c.inodeTableOverhead = c.inodesPerGroup / inodeEntriesPerBlock

	c.groupOverhead = 2 + c.inodeTableOverhead + 1 + c.blocksForBGDT

}

func (c *constants) computeBlocksPerBlockGroup() {

	// divide available blocks amongst block groups
	c.blocksPerGroup = uint32(ceiling(int64(c.totalBlocks), int64(c.totalGroups)))

	// note: removing overhead from available blocks should be uneccesary
	// as we always use the bare minimum number of block groups.

}
