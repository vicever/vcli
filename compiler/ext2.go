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
package compiler

// TODO: clean this file

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/sisatech/vcli/compiler/vmdk"
)

// CreateISO9660FS create a filesystem on the second partition of the vmdk
func createExtFS(file *os.File, auxFolder string, startWrite int, maxSectors int) int {

	if len(auxFolder) == 0 {
		return 0
	}

	_, err := os.Stat(auxFolder)
	if err != nil {
		fmt.Printf("Auxiliary folder %s does not exist.", auxFolder)
		os.Exit(1)
	}

	ext2File, err := ioutil.TempFile("", "ext2")
	if err != nil {
		fmt.Printf("Could not create filesystem: %s\n", err.Error())
		os.Exit(1)
	}

	size, err := dirSize(auxFolder)
	if err != nil {
		fmt.Printf("Could not create filesystem1: %s\n", err.Error())
		os.Exit(1)
	}

	// round to sector size
	size = (int64(vmdk.SectorSize) - size%int64(vmdk.SectorSize)) + size

	if size == 0 {
		return 0
	}

	blocks := size / int64(vmdk.SectorSize)
	if blocks < 10 {
		blocks = 10
	}

	fmt.Printf("BLOCKS: %d/%d\n", blocks, maxSectors)

	command := exec.Command("genext2fs", "-b", strconv.Itoa(int(blocks)), "-d", auxFolder, ext2File.Name())
	_, err = command.CombinedOutput()

	if err != nil {
		fmt.Printf("Could not create filesystem: %s\n", err.Error())
		os.Exit(1)
	}

	// fmt.Printf("Creating filesystem at %s\n", ext2File.Name())

	// fmt.Printf("Writing filesystem to %s\n", file.Name())

	// write the file to the vmdk
	// TODO super dodgy. same method is in vmdk.go
	app, err := os.Open(ext2File.Name())

	if err != nil {
		fmt.Printf("Could not create filesystem: %s\n", err.Error())
		os.Exit(1)
	}

	defer app.Close()

	if err != nil {
		fmt.Printf("Could not create filesystem: %s\n", err.Error())
		os.Exit(1)
	}

	a, err := ioutil.ReadAll(app)
	if err != nil {
		fmt.Printf("Could not create filesystem: %s\n", err.Error())
		os.Exit(1)
	}

	_, err = file.WriteAt(a, int64(startWrite))
	if err != nil {
		fmt.Printf("Could not create filesystem: %s\n", err.Error())
		os.Exit(1)
	}

	// get size of the new file
	info, err := ext2File.Stat()
	if err != nil {
		fmt.Printf("Could not create filesystem: %s\n", err.Error())
		os.Exit(1)
	}

	if int(info.Size())/int(vmdk.SectorSize) > maxSectors {
		fmt.Println("Auxiliary directory too big for specified disk size.")
		os.Exit(1)
	}

	fsWrite := info.Size()
	fsWrite = int64(ceiling(uint64(fsWrite), uint64(vmdk.SectorSize)) * vmdk.SectorSize)

	// TODO needs to be change to int64
	return int(fsWrite)

}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
