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

import (
	"encoding/binary"
	"errors"
	"os"
	"unsafe"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/compiler/vmdk"
)

// ReadAppNameFromVMDK reads the app name stored within a Vorteil VMDK file.
func ReadAppNameFromVMDK(filepath string) (string, error) {
	var err error
	var name string

	err = sherlock.Try(func() {

		f, e := os.Open(filepath)
		sherlock.Check(e)
		defer f.Close()
		sherlock.Check(f.Seek(int64(64), 0))
		var overhead uint64
		sherlock.Check(binary.Read(f, binary.LittleEndian, &overhead))
		overhead = vmdk.SectorSize*overhead + (512 * 34)
		imageHeader := &vmdk.ImageHeader{}
		offset := unsafe.Offsetof(imageHeader.Name)
		sherlock.Check(f.Close())
		f, e = os.Open(filepath)
		sherlock.Check(e)
		sherlock.Check(f.Seek(int64(overhead+uint64(offset)), 0))
		var nameBytes = make([]byte, 64, 64)
		sherlock.Check(f.Read(nameBytes))
		for i, c := range nameBytes {
			if c == 0 {
				name = string(nameBytes[:i])
				break
			}
		}
		if name == "" {
			sherlock.Throw(errors.New("bad name string in vmdk"))
		}

	})

	return name, err
}

// ReadNetworkCardCountFromVMDK reads the number of required network cards from
// within a Vorteil VMDK file.
func ReadNetworkCardCountFromVMDK(filepath string) (int, error) {
	var err error
	var count int

	err = sherlock.Try(func() {

		f, e := os.Open(filepath)
		sherlock.Check(e)
		defer f.Close()
		sherlock.Check(f.Seek(int64(64), 0))
		var overhead uint64
		sherlock.Check(binary.Read(f, binary.LittleEndian, &overhead))
		overhead = vmdk.SectorSize*overhead + (512 * 34)
		imageHeader := &vmdk.ImageHeader{}
		offset := unsafe.Offsetof(imageHeader.Cards)
		sherlock.Check(f.Seek(int64(overhead+uint64(offset)), 0))
		var cardBytes = make([]byte, 192, 192)
		for i := 0; i < 4; i++ {
			sherlock.Check(f.Read(cardBytes))
			for _, c := range cardBytes {
				if c != 0 {
					count++
					break
				}
			}
		}

	})

	return count, err
}
