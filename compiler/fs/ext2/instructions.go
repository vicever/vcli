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
	"io"
	"os"
)

type instructions []*instruction

func (ins instructions) Len() int {
	return len(ins)
}

func (ins instructions) Less(i, j int) bool {
	return ins[i].offset < ins[j].offset
}

func (ins instructions) Swap(i, j int) {
	tmp := ins[i]
	ins[i] = ins[j]
	ins[j] = tmp
}

type Instructions struct {
	index int
	*instruction
	instructions []*instruction
	constants
	inodes    uint32
	blocks    uint32
	nodes     []*Inode
	groupDirs []int
}

type instruction struct {
	file    *os.File
	offset  int64
	length  int64
	fOffset int64
	fPath   string
	data    *bytes.Buffer
}

func (ins *Instructions) Next() bool {

	var err error

	if ins.instruction != nil {
		if ins.file != nil {
			ins.file.Close()
		}
	}

	ins.index++

	if ins.index < len(ins.instructions) {

		ins.instruction = ins.instructions[ins.index]

		if ins.fPath != "" {

			ins.file, err = os.Open(ins.fPath)
			if err != nil {
				panic(err)
			}

			_, err = ins.file.Seek(ins.fOffset, 0)
			if err != nil {
				panic(err)
			}

		}

		return true

	}

	return false

}

func (ins *Instructions) Offset() int64 {

	return ins.offset

}

func (ins *Instructions) Length() int64 {

	return ins.length

}

func (ins *Instructions) Data() io.Reader {

	if ins.file == nil {
		return ins.data
	}

	return ins.file

}
