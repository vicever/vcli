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
	"io/ioutil"
	"os"
)

func (build *builder) writeTrampoline() error {

	path := build.env.path + "/vtramp.img"

	if _, err := os.Stat(path); os.IsNotExist(err) {

		// TODO
		// err = DownloadVorteilFile("vtramp.img", "vtramp")
		if err != nil {
			return err
		}

	}

	trampoline, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = build.content.file.WriteAt(trampoline, int64(build.content.trampoline.first*SectorSize))
	if err != nil {
		return err
	}

	build.log("Writing trampoline at LBAs: %d - %d", build.content.trampoline.first, build.content.trampoline.last)

	return nil

}
