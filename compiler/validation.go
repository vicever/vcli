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
	"debug/elf"
	"encoding/json"
	"errors"
	"fmt"
	"image/png"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sisatech/vcli/shared"
)

// FullValidation checks all components needed to build a Vorteil disk image.
func FullValidation(binary, config, files, icon, kernel string) error {

	if kernel == "" {
		return errors.New("must specify a kernel")
	}

	if s := strings.Split(kernel, "."); len(s) != 3 {
		return errors.New("invalid kernel argument")
	}

	err := ValidateVCFG(config)
	if err != nil {
		return err
	}

	if files != "" {

		// err = ValidateFiles(files, config)
		// if err != nil {
		// 	return err
		// }

	}

	if icon != "" {

		err = ValidateIcon(icon)
		if err != nil {
			return err
		}

	}

	return nil

}

// ValidateELF checks if the target is a supported ELF binary file.
func ValidateELF(target string) error {

	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return errors.New("target is a directory, not an ELF binary")
	}

	f, err := elf.Open(target)
	if err != nil {
		return errors.New("target is not an ELF binary")
	}

	f.Close()

	return nil

}

// ValidateVCFG checks if the file parses to a valid build configuration file.
func ValidateVCFG(target string) error {

	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return errors.New("target is a directory, not an ELF binary")
	}

	in, err := ioutil.ReadFile(target)
	if err != nil {
		return err
	}

	vcfg := new(shared.BuildConfig)
	err = json.Unmarshal(in, vcfg)
	// err = yaml.Unmarshal(in, vcfg)
	if err != nil {
		return errors.New("vcfg file failed to parse correctly")
	}

	if vcfg.App == nil || vcfg.Network == nil || vcfg.Disk == nil {
		return errors.New("target is invalid, incomplete, or outdated")
	}

	if vcfg.Disk.DiskSize == 0 {
		return errors.New("disk size set to 0 in config file")
	}

	return nil

}

// ValidateIcon checks if the target is a valid .png file.
func ValidateIcon(target string) error {

	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return errors.New("target is a directory, not an ELF binary")
	}

	f, err := os.Open(target)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = png.DecodeConfig(f)
	if err != nil {
		return err
	}

	return nil

}

// ValidateFiles checks if the target is a directory and fits within the number
// of megabytes defined by the config file. The config argument can be left
// empty to simply validate the target without concern for size.
func ValidateFiles(target, config string) error {
	size := 0

	// determine size from config file
	if config != "" {

		in, err := ioutil.ReadFile(config)
		if err != nil {
			return err
		}

		vcfg := new(shared.BuildConfig)
		err = json.Unmarshal(in, vcfg)
		// err = yaml.Unmarshal(in, vcfg)
		if err != nil {
			return err
		}

		size = vcfg.Disk.DiskSize

	}

	// validate folder
	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New("target is not a directory")
	}

	// compute size of folder
	if size > 0 {

		fmt.Printf("SIZE: %d\n", size)

		dirSize, err := recurseDirSize(target)
		if err != nil {
			return err
		}

		// convert bytes to megabytes
		dirSize = dirSize / 0x100000
		if int(dirSize) <= size {
			return errors.New("files tree too large to fit on disk")
		}

	}

	return nil

}

func recurseDirSize(target string) (int64, error) {

	var size int64

	info, err := os.Stat(target)
	if err != nil {
		return 0, err
	}

	if info.IsDir() {

		infos, err := ioutil.ReadDir(target)
		if err != nil {
			return 0, err
		}

		for _, x := range infos {

			n, err := recurseDirSize(x.Name())
			if err != nil {
				return 0, err
			}

			size = size + n

		}

	} else {

		return info.Size(), nil

	}

	return size, nil

}
