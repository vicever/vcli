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
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sisatech/vcli/shared"
)

type builder struct {
	env          *Environment
	disk         *os.File
	files        *os.File
	args         *BuildArgs
	config       *shared.BuildConfig
	descriptor   *bytes.Buffer
	overhead     overhead
	content      content
	header       Header
	seek         int64
	sector       uint64
	grainCounter uint64
	grains       []uint64
	finalGrain   []byte
}

func (env *Environment) newBuilder(args *BuildArgs) *builder {

	build := &builder{
		env:  env,
		args: args,
	}

	return build

}

func (build *builder) log(s string, args ...interface{}) {

	if build.env.talk {
		fmt.Printf(s+"\n", args...)
	}

}

func (build *builder) newDisk() error {

	var err error

	if build.args.Destination == "" {

		// create as a temp file

		build.disk, err = ioutil.TempFile("", "vmdk-")
		if err != nil {
			return err
		}

	} else {

		// create file at destination

		// check if a file already exists
		var info os.FileInfo
		info, err = os.Stat(build.args.Destination)
		if !os.IsNotExist(err) {

			// delete any existing file but not a directory
			if info != nil && info.IsDir() {
				return fmt.Errorf("destination '%s' is a directory",
					build.args.Destination)
			}

			err = os.Remove(build.args.Destination)
			if err != nil {
				return err
			}

		}

		// open new file
		build.disk, err = os.Create(build.args.Destination)
		if err != nil {
			return err
		}

	}

	build.log("Disk: %s", build.disk.Name())

	return nil

}
