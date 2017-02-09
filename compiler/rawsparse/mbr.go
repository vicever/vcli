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
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
)

type protectiveMBREntry struct {
	status          byte
	headFirst       byte
	sectorFirst     byte
	cylinderFirst   byte
	partitionType   byte
	headLast        byte
	sectorLast      byte
	cylinderLast    byte
	firstLBA        uint32
	numberOfSectors uint32
}

func (build *builder) writeMBR() error {

	var err error

	// write bootloader at start of sector
	path := fmt.Sprintf("%s/%s", build.env.path, "vboot.img")
	build.log("MBR: %s", path)

	if _, err = os.Stat(path); os.IsNotExist(err) {

		// TODO
		// err = DownloadVorteilFile("vboot.img", "vboot")
		if err != nil {
			return err
		}

	}

	mbr, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = build.content.file.Write(mbr)
	if err != nil {
		return err
	}

	// write protective MBR entry
	protMBR := &protectiveMBREntry{
		status:          0x7f,
		headFirst:       0,
		sectorFirst:     0,
		cylinderFirst:   0,
		partitionType:   0xEE,
		headLast:        0,
		sectorLast:      0,
		cylinderLast:    0,
		firstLBA:        1,
		numberOfSectors: uint32(build.content.LBAs),
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, protMBR)
	if err != nil {
		return err
	}

	_, err = build.content.file.WriteAt(buf.Bytes(), 446)
	if err != nil {
		return err
	}

	// write magic number at end of sector
	_, err = build.content.file.WriteAt([]byte{0x55, 0xAA}, 510)
	if err != nil {
		return err
	}

	return nil

}
