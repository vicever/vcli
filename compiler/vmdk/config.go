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
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	argsMax = 16
	argsLen = 64

	dnsMax = 4
	dnsLen = 64

	envsMax = 16
	envsLen = 64
)

// ImageHeaderCard ...
type ImageHeaderCard struct {
	ip   [64]byte
	mask [64]byte
	gw   [64]byte
}

// Redirect
type Redirect struct {
	Src      [64]byte
	Dest     [64]byte
	Protocol [64]byte
}

// ImageHeader ...
// we write partition length and start into config later we need to adjust the
// offset in writeConfig, if we add stuff before those entries
type ImageHeader struct {
	lbaKernelStart     uint32
	lbaKernelLength    uint32
	lbaAppStart        uint32
	lbaAppLength       uint32
	lbaTrampStart      uint32
	lbaPartitionStart  uint32
	lbaPartitionLength uint32
	Name               [64]byte
	args               [16][64]byte
	envs               [16][64]byte
	dns                [4][64]byte
	Cards              [4]ImageHeaderCard
	fsType             [64]byte
	maxFd              uint32
	tsHost             [64]byte
	tsServers          [4][64]byte
	fileRedirects      [4]Redirect
	randBytes          [16]byte
}

func randomByteGen() ([16]byte, error) {
	var v [16]byte
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return v, err
	}
	for i, z := range b {
		v[i] = z
	}
	return v, nil
}

func (build *builder) writeConfig() error {

	var args [argsMax][argsLen]byte
	var dns [dnsMax][dnsLen]byte
	var envs [envsMax][envsLen]byte
	var ntpServers [4][64]byte

	convertStringToBytes := func(in string, out []byte) {
		copy(out, in)
	}

	// time Servers
	var ntpHost [64]byte
	convertStringToBytes(build.config.NTP.Hostname, ntpHost[:])
	// fmt.Printf("%v\n", len(build.config.NTP.Servers))
	for i, element := range build.config.NTP.Servers {
		convertStringToBytes(element, ntpServers[i][:])
	}

	// file redirects
	var fileRedirects [4]Redirect
	for i, element := range build.config.Redirects.Rules {
		rule := new(Redirect)
		convertStringToBytes(element.Dest, rule.Dest[:])
		convertStringToBytes(element.Src, rule.Src[:])
		convertStringToBytes(element.Protocol, rule.Protocol[:])
		fileRedirects[i] = *rule
	}

	// program args
	if len(build.config.App.BinaryArgs) > argsMax {
		return fmt.Errorf("too many program arguments: %d; maximum %d", len(build.config.App.BinaryArgs), argsMax)
	}

	for i, element := range build.config.App.BinaryArgs {

		if len(element) > argsLen {
			return fmt.Errorf("program argument too long: `%s`; maximum %d", element, argsLen)
		}

		convertStringToBytes(element, args[i][:])

	}

	// environment variables
	if len(build.config.App.SystemEnvs) > envsMax {
		return fmt.Errorf("too many environment variables: %d; maximum %d", len(build.config.App.SystemEnvs), envsMax)
	}

	for i, element := range build.config.App.SystemEnvs {

		if len(element) > envsLen {
			return fmt.Errorf("environment variable too long: `%s`; maximum %d", element, envsLen)
		}

		convertStringToBytes(element, envs[i][:])

	}

	// dns
	if len(build.config.Network.DNS) > dnsMax {
		return fmt.Errorf("too many DNS variables: %d; maximum %d", len(build.config.Network.DNS), dnsMax)
	}

	for i, element := range build.config.Network.DNS {

		if len(element) > dnsLen {
			return fmt.Errorf("DNS IP too long: %d; maximum %d", len(element), dnsLen)
		}

		convertStringToBytes(element, dns[i][:])

	}

	// network cards
	var cards [4]ImageHeaderCard
	for i, element := range build.config.Network.NetworkCards {
		card := new(ImageHeaderCard)
		convertStringToBytes(element.IP, card.ip[:])
		convertStringToBytes(element.Mask, card.mask[:])
		convertStringToBytes(element.Gateway, card.gw[:])
		cards[i] = *card
	}

	// image header
	ih := &ImageHeader{
		lbaKernelStart:     uint32(build.content.kernel.first),
		lbaKernelLength:    uint32(build.content.kernel.length + 32), // TODO: fix this! This '+32' exists due to a bug in the kernel!!!
		lbaAppStart:        uint32(build.content.app.first),
		lbaAppLength:       uint32(build.content.app.length),
		lbaTrampStart:      uint32(build.content.trampoline.first),
		lbaPartitionStart:  uint32(build.content.files.first),
		lbaPartitionLength: uint32(build.content.files.length),
		args:               args,
		envs:               envs,
		dns:                dns,
		Cards:              cards,
		maxFd:              uint32(build.config.Disk.MaxFD),
		tsHost:             ntpHost,
		tsServers:          ntpServers,
		fileRedirects:      fileRedirects,
	}

	ih.randBytes, _ = randomByteGen()

	convertStringToBytes(build.config.Name, ih.Name[:])
	convertStringToBytes(build.config.Disk.FileSystem, ih.fsType[:])

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, ih)
	if err != nil {
		return fmt.Errorf("failed to write image header: %s", err.Error())
	}

	_, err = build.content.file.WriteAt(buf.Bytes(), int64(build.content.config.first*SectorSize))
	if err != nil {
		return fmt.Errorf("failed to write image header: %s", err.Error())
	}

	return nil

}
