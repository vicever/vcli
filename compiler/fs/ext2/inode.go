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

const (
	inodeDirectoryPermissions = 0x4000 | 0x1FF
	inodeFilePermissions      = 0x8000 | 0x1FF
)

type Inode struct {
	Permissions      uint16
	UID              uint16
	SizeLower        uint32
	LastAccessTime   uint32
	CreationTime     uint32
	ModificationTime uint32
	DeletionTime     uint32
	GID              uint16
	Links            uint16
	Sectors          uint32
	Flags            uint32
	OSV              uint32
	DirectPointer    [12]uint32
	SinglyIndirect   uint32
	DoublyIndirect   uint32
	TriplyIndirect   uint32
	GenNo            uint32
	Reserved         [2]uint32
	FragAddr         uint32
	OSStuff          [12]byte
}

func (ins *Instructions) inodePointers(start, length uint32, inode *Inode) {

	// direct pointers
	for i := uint32(0); i < 12; i++ {

		if length <= i {
			return
		}

		inode.DirectPointer[i] = ins.mapDBtoBlockAddr(start + i)

	}

	// singly indirect
	if length > 12 {
		inode.SinglyIndirect = ins.mapDBtoBlockAddr(start + 12)
	} else {
		return
	}

	// doubly indirect
	if length > 256+12+1 {
		inode.DoublyIndirect = ins.mapDBtoBlockAddr(start + 256 + 12 + 1)
	} else {
		return
	}

	// triply indirect
	if length > 12+1+256+1+256+256*256 {
		inode.TriplyIndirect = ins.mapDBtoBlockAddr(start + 12 + 1 + 256 + 1 + 256 + 256*256)
	}

}

func (ins *Instructions) mapDBtoBlockAddr(in uint32) uint32 {

	grp := in / (ins.blocksPerGroup - ins.groupOverhead)
	off := in % (ins.blocksPerGroup - ins.groupOverhead)

	return ins.overhead + grp*ins.blocksPerGroup + ins.groupOverhead + off

}
