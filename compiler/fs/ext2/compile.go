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
	"fmt"
	"sort"

	"github.com/sisatech/sherlock"
)

func Compile(path string, blocks, inodes uint32) (*Instructions, error) {

	ins := new(Instructions)
	ins.inodes = 10
	ins.compute(blocks, inodes)

	ins.groupDirs = make([]int, ins.totalGroups)

	err := sherlock.Try(func() {

		ins.scanRoot(path)

		ins.writeSuperblock()
		ins.writeBGDT()

		ins.writeBlockGroups()

		// TODO fix last byte
		ins.instructions = append(ins.instructions, &instruction{
			offset: int64(blocks*blockSize) - 1,
			length: 1,
			data:   bytes.NewBufferString(" "),
		})

		ins.index = -1

	})

	if ins.blocks > blocks {
		additional := ceiling(int64(ins.blocks)-int64(blocks), blockSize)
		return nil, fmt.Errorf("disk size not big enough for filesystem; needs to be at least %v MiB bigger", additional)
	}

	sort.Sort(instructions(ins.instructions))

	return ins, err

}
