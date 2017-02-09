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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

func (build *builder) writeKernel() error {

	debug := "PROD"
	if build.args.Debug {
		debug = "DEBUG"
	}

	kernel := fmt.Sprintf("vkernel-%s-%s.img", debug, build.args.Kernel)
	path := fmt.Sprintf("%s/%s", build.env.path, kernel)

	if _, err := os.Stat(path); os.IsNotExist(err) {

		// TODO try to download kernel
		// build.Log("%s not found locally. Trying to download.", kernel)
		// err := DownloadVorteilFile(kernel, "vkernel")
		if err != nil {
			return err
		}

	}

	build.log("Kernel: %s", path)

	in, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New("error reading kernel: " + err.Error())
	}

	_, err = build.content.file.WriteAt(in, int64(build.content.kernel.first*SectorSize))
	if err != nil {
		return errors.New("error writing kernel: " + err.Error())
	}

	build.log("Writing kernel at LBAs: %d - %d", build.content.kernel.first, build.content.kernel.last)

	return nil

}
