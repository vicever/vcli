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
package vmdk

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
)

const (
	partitionArraySize = 128 * 128
)

var (
	partitionOneName = [72]byte{0x76, 0x0, 0x6f, 0x0, 0x72, 0x0, 0x74, 0x0,
		0x65, 0x0, 0x69, 0x0, 0x6c, 0x0, 0x2d, 0x0, 0x6f, 0x0, 0x73, 0x0}

	partitionTwoName = [72]byte{0x76, 0x0, 0x6f, 0x0, 0x72, 0x0, 0x74, 0x0,
		0x65, 0x0, 0x69, 0x0, 0x6c, 0x0, 0x2d, 0x0, 0x72, 0x0, 0x6f,
		0x0, 0x6f, 0x0, 0x74}
)

type gptHeader struct {
	signature      uint64
	revision       [4]byte
	headerSize     uint32
	crc            uint32
	zero           uint32
	currentLBA     uint64
	backupLBA      uint64
	firstUsableLBA uint64
	lastUsableLBA  uint64
	guid           [16]byte
	startLBAParts  uint64
	noOfParts      uint32
	sizePartEntry  uint32
	crcParts       uint32
}

type gptPartition struct {
	typeGUID   [16]byte
	partGUID   [16]byte
	firstLBA   uint64
	lastLBA    uint64
	attributes uint64
	name       [72]byte
}

func generateUID() [16]byte {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(crand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return [16]byte{}
	}
	uuid[8] = uuid[8]&^0xc0 | 0x80
	uuid[6] = uuid[6]&^0xf0 | 0x40

	var retArr [16]byte
	copy(retArr[:], uuid[0:15])
	return retArr
}

func hashFilePart(file *os.File, start, length int) uint32 {

	arrayHash := crc32.NewIEEE()
	bin := make([]byte, length)
	file.ReadAt(bin, 2*SectorSize)
	arrayHash.Write(bin)

	return arrayHash.Sum32()

}

func hashGPT(gpt *gptHeader) (uint32, error) {

	buf := new(bytes.Buffer)
	arrayHash := crc32.NewIEEE()
	err := binary.Write(buf, binary.LittleEndian, gpt)
	if err != nil {
		return 0, err
	}
	arrayHash.Write(buf.Bytes())

	return arrayHash.Sum32(), nil

}

func (build *builder) writeGPT() error {

	var err error

	gpt := &gptHeader{
		signature:      0x5452415020494645,
		revision:       [4]byte{0, 0, 1, 0},
		headerSize:     92,
		crc:            0,
		zero:           0,
		currentLBA:     1,
		backupLBA:      uint64(build.content.LBAs - 1),
		firstUsableLBA: build.content.reserved.last + 1,
		lastUsableLBA:  build.content.LBAs - build.content.reserved.length,
		guid:           generateUID(),
		startLBAParts:  2,
		noOfParts:      128,
		sizePartEntry:  128,
		crcParts:       0,
	}

	// partition 1
	p1Buf := new(bytes.Buffer)
	partVorteil := &gptPartition{
		typeGUID: [16]byte{},
		partGUID: generateUID(),
		firstLBA: build.content.reserved.last + 1,
		lastLBA:  build.content.app.last,
		name:     partitionOneName,
	}

	err = binary.Write(p1Buf, binary.LittleEndian, partVorteil)
	_, err = build.content.file.WriteAt(p1Buf.Bytes(), 2*SectorSize)
	if err != nil {
		return err
	}

	// partition 2
	p2Buf := new(bytes.Buffer)
	partRoot := &gptPartition{
		typeGUID: [16]byte{0xB6, 0x7C, 0x6E, 0x51, 0xCF, 0x6E,
			0xD6, 0x11, 0x8F, 0xF8, 0x00, 0x02, 0x2D, 0x09, 0x71, 0x2B},
		partGUID: generateUID(),
		firstLBA: build.content.files.first,
		lastLBA:  build.content.files.last,
		name:     partitionTwoName,
	}

	// checksum the partition table
	err = binary.Write(p2Buf, binary.LittleEndian, partRoot)
	_, err = build.content.file.WriteAt(p2Buf.Bytes(), 2*SectorSize+128)
	gpt.crcParts = hashFilePart(build.content.file, 2*SectorSize, partitionArraySize)
	if err != nil {
		return err
	}

	// checksum the header
	hash, err := hashGPT(gpt)
	gpt.crc = hash

	// write gpt to file
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, gpt)
	_, err = build.content.file.WriteAt(buf.Bytes(), SectorSize)
	if err != nil {
		return err
	}

	// create secondary GPT for later use

	build.finalGrain = make([]byte, sectorsPerGrain*SectorSize, sectorsPerGrain*SectorSize)

	copy(build.finalGrain[(sectorsPerGrain-33)*SectorSize:], p1Buf.Bytes())
	copy(build.finalGrain[(sectorsPerGrain-33)*SectorSize+128:], p2Buf.Bytes())

	gpt.backupLBA = 1
	gpt.currentLBA = build.content.backup.last
	gpt.startLBAParts = build.content.backup.first + 1
	gpt.crc = 0

	hash, err = hashGPT(gpt)
	gpt.crc = hash

	buf = new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, gpt)

	copy(build.finalGrain[(sectorsPerGrain-1)*SectorSize:], buf.Bytes())

	return nil

}
