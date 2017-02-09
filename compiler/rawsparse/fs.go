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
package rawsparse

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/sisatech/vcli/compiler/fs/ext2"
)

func CompileFilesystem(path string, sectors uint64) (*os.File, error) {

	ext2File, err := ioutil.TempFile("", "ext2")
	if err != nil {
		return nil, err
	}
	defer ext2File.Close()

	ins, err := ext2.Compile(path, uint32(sectors/2), 4096)
	if err != nil {
		return nil, err
	}

	for ins.Next() {
		_, err = ext2File.Seek(ins.Offset(), 0)
		if err != nil {
			return nil, err
		}

		_, err = io.CopyN(ext2File, ins.Data(), ins.Length())
		if err != nil {
			return nil, err
		}
	}

	// command := exec.Command("genext2fs", "-b", strconv.Itoa(int(sectors/2)), "-d", path, ext2File.Name())
	//
	// out, err := command.CombinedOutput()
	// if err != nil {
	// 	if strings.HasPrefix(string(out), "genext2fs: couldn't allocate a block (no free space)") ||
	// 		strings.HasPrefix(string(out), "genext2fs: reserved blocks value is invalid.") {
	// 		return nil, fmt.Errorf("disk size specified in vcfg file not large enough to fit the files")
	// 	}
	// 	fmt.Println(string(out))
	// 	return nil, err
	// }

	return ext2File, nil

}

func (build *builder) writeFilesystem() error {

	path := build.files.Name()

	fs, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fs.Close()

	in, err := ioutil.ReadAll(fs)
	if err != nil {
		return err
	}

	_, err = build.content.file.WriteAt(in, int64(build.content.files.first*SectorSize))
	if err != nil {
		return err
	}

	build.log("Writing filesystem at LBAs: %d - %d", build.content.files.first, build.content.files.last)

	return nil

}
