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
	"errors"
	"io/ioutil"
	"os"

	"github.com/sisatech/sherlock"
)

const (
	pointerSize = 4

	maxDirectPointers = 12

	dirNameAlignment = 4

	sectorSize = 512

	sectorsPerBlock = blockSize / sectorSize
)

func (ins *Instructions) scanRoot(path string) {

	// reserved inode 1
	ins.nodes = append(ins.nodes, &Inode{})

	var dataLength int64

	children, err := ioutil.ReadDir(path)
	sherlock.Check(err)

	dataLength = dirDataLength(children)
	dataBlocks := ceiling(dataLength, blockSize)
	blocks := computeBlocks(dataLength)
	start := ins.blocks
	ins.blocks += blocks

	dirChildren := uint32(0)
	for _, child := range children {
		if child.IsDir() {
			dirChildren++
		}
	}

	// reserved inode 2 (root inode)
	inode := &Inode{
		Permissions:      inodeDirectoryPermissions,
		UID:              superUID,
		SizeLower:        uint32(dataBlocks * blockSize),
		LastAccessTime:   uint32(ins.timestamp.Unix()),
		CreationTime:     uint32(ins.timestamp.Unix()),
		ModificationTime: uint32(ins.timestamp.Unix()),
		GID:              superGID,
		Links:            uint16(2 + dirChildren), //...
		Sectors:          blocks * sectorsPerBlock,
	}

	ins.inodePointers(ins.blocks-blocks, blocks, inode)

	ins.nodes = append(ins.nodes, inode)

	ins.groupDirs[0]++

	// reserved inodes 3-10
	for i := 0; i < 8; i++ {
		ins.nodes = append(ins.nodes, &Inode{})
	}

	this := uint32(2)
	parent := this

	var tuples []*dirTuple
	tuples = append(tuples, &dirTuple{name: ".", inode: this})
	tuples = append(tuples, &dirTuple{name: "..", inode: parent})

	for _, child := range children {
		tuples = append(tuples, ins.scan(path+"/"+child.Name(), this))
	}

	ins.writeDir(dirData(tuples), start, blocks)

}

type dirTuple struct {
	name  string
	inode uint32
}

func (ins *Instructions) scan(path string, parent uint32) *dirTuple {

	fi, err := os.Stat(path)
	sherlock.Check(err)

	ins.inodes++
	this := ins.inodes

	var dataLength int64

	if fi.IsDir() {

		children, err := ioutil.ReadDir(path)
		sherlock.Check(err)

		dataLength = dirDataLength(children)
		dataBlocks := ceiling(dataLength, blockSize)
		blocks := computeBlocks(dataLength)
		start := ins.blocks
		ins.blocks += blocks

		dirChildren := uint32(0)
		for _, child := range children {
			if child.IsDir() {
				dirChildren++
			}
		}

		// inode
		inode := &Inode{
			Permissions:      inodeDirectoryPermissions,
			UID:              superUID,
			SizeLower:        uint32(dataBlocks * blockSize),
			LastAccessTime:   uint32(ins.timestamp.Unix()),
			CreationTime:     uint32(ins.timestamp.Unix()),
			ModificationTime: uint32(ins.timestamp.Unix()),
			GID:              superGID,
			Links:            uint16(1 + dirChildren),
			Sectors:          blocks * sectorsPerBlock,
		}

		ins.inodePointers(ins.blocks-blocks, blocks, inode)

		ins.groupDirs[(this-1)/ins.inodesPerGroup]++

		ins.nodes = append(ins.nodes, inode)

		var tuples []*dirTuple
		tuples = append(tuples, &dirTuple{name: ".", inode: this})
		tuples = append(tuples, &dirTuple{name: "..", inode: parent})

		for _, child := range children {
			tuples = append(tuples, ins.scan(path+"/"+child.Name(), this))
		}

		ins.writeDir(dirData(tuples), start, blocks)

	} else {

		dataLength = fi.Size()
		// dataBlocks := ceiling(dataLength, blockSize)
		blocks := computeBlocks(dataLength)
		start := ins.blocks
		ins.blocks += blocks

		ins.writeFile(path, start, blocks, fi.Size())

		// inode
		inode := &Inode{
			Permissions:      inodeFilePermissions,
			UID:              superUID,
			SizeLower:        uint32(dataLength),
			LastAccessTime:   uint32(ins.timestamp.Unix()),
			CreationTime:     uint32(ins.timestamp.Unix()),
			ModificationTime: uint32(ins.timestamp.Unix()),
			GID:              superGID,
			Links:            uint16(1),
			Sectors:          blocks * sectorsPerBlock,
		}

		ins.inodePointers(ins.blocks-blocks, blocks, inode)

		ins.nodes = append(ins.nodes, inode)

	}

	return &dirTuple{name: fi.Name(), inode: this}

}

func (ins *Instructions) writeDir(buf *bytes.Buffer, start, length uint32) {

	index := 0

	for i := uint32(0); i < length; i++ {

		addr := ins.mapDBtoBlockAddr(i + start)

		// if single indirect block
		if i == 12 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < 256+i+1; j++ {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			break

		}

		// if double indirect first-level block
		if i == 269 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < i+1+257*256; j = j + 257 {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			break

		}

		// if double indirect second-level block
		if i > 269 && i < 269+256*257 && (i-269)%257 == 0 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < 256+i+1; j++ {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			break

		}

		// if triple indirect first-level block
		if i == 269+2+256+256*256 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < i+1+257*256+256*256*256; j = j + 257*256 {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			break

		}

		// if triple indirect second-level block
		if i > 269+2+256+256*256 && (i-269+2+256+256*256)%(257*256) == 0 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < i+1+257*256; j = j + 257 {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			break

		}

		// if triple indirect third-level block
		if i > 269+2+256+256*256 && (1+(i-269+2+256+256*256)%(257*256))%257 == 0 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < 256+i+1; j++ {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			break

		}

		// add block
		ins.instructions = append(ins.instructions, &instruction{
			offset: int64(addr * blockSize),
			length: blockSize,
			data:   bytes.NewBuffer(buf.Bytes()[index*blockSize : (index+1)*blockSize]),
		})

		index++

	}

}

func (ins *Instructions) writeFile(path string, start, length uint32, actual int64) {

	index := 0

	for i := uint32(0); i < length; i++ {

		addr := ins.mapDBtoBlockAddr(i + start)

		// if single indirect block
		if i == 12 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < 256+i+1; j++ {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			continue

		}

		// if double indirect first-level block
		if i == 269 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < i+1+257*256; j = j + 257 {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			continue

		}

		// if double indirect second-level block
		if i > 269 && i < 269+256*257 && (i-269-1)%257 == 0 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < 256+i+1; j++ {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			continue

		}

		// if triple indirect first-level block
		if i == 269+1+256+256*256 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < i+1+257*256+256*256*256; j = j + 257*256 {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			continue

		}

		// if triple indirect second-level block
		if i > 269+1+256+256*256 && (i-(269+2+256+256*256))%(257*256+1) == 0 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < i+1+257*256; j = j + 257 {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			continue

		}

		// if triple indirect third-level block
		if i > 269+2+256+256*256 && ((i-(269+2+256+256*256))%(257*256+1)-1)%257 == 0 {

			buffer := new(bytes.Buffer)
			for j := i + 1; j < length && j < 256+i+1; j++ {
				sherlock.Check(binary.Write(buffer, binary.LittleEndian, uint32(ins.mapDBtoBlockAddr(j+start))))
			}

			// add block
			ins.instructions = append(ins.instructions, &instruction{
				offset: int64(addr * blockSize),
				length: int64(buffer.Len()),
				data:   buffer,
			})

			continue

		}

		l := int64(blockSize)

		if i == length-1 && actual%blockSize != 0 {
			l = actual % blockSize
		}

		// add block
		ins.instructions = append(ins.instructions, &instruction{
			offset:  int64(addr * blockSize),
			length:  l,
			fPath:   path,
			fOffset: int64(index * blockSize),
		})

		index++

	}

}

func dirData(tuples []*dirTuple) *bytes.Buffer {

	buf := new(bytes.Buffer)

	length := int64(0)
	leftover := int64(blockSize)

	for i, child := range tuples {

		l := 8 + align(int64(len(child.name)+1), dirNameAlignment)

		if leftover >= l {

			length += l
			leftover -= l

		} else {

			// add a null entry into the leftover space
			// inode
			err := binary.Write(buf, binary.LittleEndian, uint32(0))
			sherlock.Check(err)

			// entry size
			err = binary.Write(buf, binary.LittleEndian, uint16(leftover))
			sherlock.Check(err)

			// name length
			err = binary.Write(buf, binary.LittleEndian, uint16(0))
			sherlock.Check(err)

			// padding
			_, err = buf.Write(bytes.Repeat([]byte{0}, int(leftover-8)))
			sherlock.Check(err)

			length += leftover
			length += l
			leftover = blockSize - l

		}

		if leftover < 8 || i == len(tuples)-1 {
			l += leftover
			length += leftover
			leftover = blockSize
		}

		// inode
		err := binary.Write(buf, binary.LittleEndian, child.inode)
		sherlock.Check(err)

		// entry size
		err = binary.Write(buf, binary.LittleEndian, uint16(l))
		sherlock.Check(err)

		// name length
		err = binary.Write(buf, binary.LittleEndian, uint16(len(child.name)))
		sherlock.Check(err)

		// name
		err = binary.Write(buf, binary.LittleEndian, append([]byte(child.name), 0))
		sherlock.Check(err)

		// padding
		_, err = buf.Write(bytes.Repeat([]byte{0}, int(l-8-int64(len(child.name))-1)))
		sherlock.Check(err)

	}

	buf.Grow(int(leftover))

	return buf

}

func dirDataLength(children []os.FileInfo) int64 {

	// '.' entry + ".." entry
	length := int64(24)
	leftover := blockSize - length

	for _, child := range children {

		l := 8 + align(int64(len(child.Name())), dirNameAlignment)

		if leftover >= l {

			length += l
			leftover -= l

		} else {

			length += leftover
			length += l
			leftover = blockSize - l

		}

	}

	return length

}

func computeBlocks(len int64) uint32 {

	x := int64(blockSize / pointerSize)

	data := ceiling(len, blockSize)

	// indirect pointer thresholds
	single := int64(maxDirectPointers)
	double := single + x
	triple := double + x*x

	// file too large threshold
	bounds := triple + x*x*x

	switch {

	case data <= single:
		return uint32(data)

	case data <= double:
		return uint32(data +
			1)

	case data <= triple:

		return uint32(data +
			1 +
			1 + ceiling(data-double, x))

	case data <= bounds:

		return uint32(data +
			1 +
			1 + x +
			1 + ceiling(ceiling(data-triple, x), x) + ceiling(data-triple, x))

	default:
		sherlock.Throw(errors.New("file too large for ext2"))
		return 0

	}

}
