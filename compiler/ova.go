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
	"archive/tar"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/sisatech/vcli/shared"
)

// BuildOVA returns the name of a compiled .ova disk image within a temporary
// folder. The caller should move the file to a non-temporary location.
func BuildOVA(binary, config, files, kernel, destination string, debug bool) (*os.File, error) {

	err := FullValidation(binary, config, files, "", kernel)
	if err != nil {
		return nil, err
	}

	in, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, err
	}

	cfg := new(shared.BuildConfig)
	err = json.Unmarshal(in, cfg)
	// err = yaml.Unmarshal(in, cfg)
	if err != nil {
		return nil, err
	}

	// TODO: implement

	// build stream optimized vmdk
	vmdk, err := BuildStreamOptimizedVMDK(binary, config, files, kernel, "", debug)
	if err != nil {
		return nil, err
	}

	// rename vmdk
	err = os.Rename(vmdk.Name(), vmdk.Name()+"-disk1.vmdk")
	if err != nil {
		return nil, err
	}

	vmdk, err = os.Open(vmdk.Name() + "-disk1.vmdk")
	if err != nil {
		return nil, err
	}

	// build ovf file
	ovf, err := generateOVF(vmdk.Name(), strconv.FormatInt(int64(cfg.Disk.DiskSize), 10))
	if err != nil {
		return nil, err
	}
	defer os.Remove(ovf.Name())
	defer os.Remove(vmdk.Name())

	// tar into ova file
	var ova *os.File
	if destination == "" {
		ova, err = ioutil.TempFile("", "")
	} else {
		ova, err = os.Create(destination)
	}
	if err != nil {
		return nil, err
	}
	defer ova.Close()

	tarrer := tar.NewWriter(ova)
	defer tarrer.Close()

	err = addToTar(tarrer, ovf)
	if err != nil {
		return ova, err
	}

	err = addToTar(tarrer, vmdk)
	if err != nil {
		return ova, err
	}

	if destination == "" {

		err = ova.Close()
		if err != nil {
			return ova, err
		}

		err = os.Rename(ova.Name(), ova.Name()+".ova")
		if err != nil {
			return ova, err
		}

		ova, err = os.Open(ova.Name() + ".ova")
		if err != nil {
			return ova, err
		}

	}

	return ova, nil

}

func addToTar(tarrer *tar.Writer, file *os.File) error {

	info, err := os.Stat(file.Name())
	if err != nil {
		return err
	}

	ti, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}

	err = tarrer.WriteHeader(ti)
	if err != nil {
		return err
	}

	_, err = io.Copy(tarrer, file)
	if err != nil {
		return err
	}

	return nil

}
