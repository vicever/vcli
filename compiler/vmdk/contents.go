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

import "io/ioutil"

func (build *builder) diskContents() error {

	var err error

	build.content.file, err = ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	defer build.content.file.Close()

	// write reserved LBAs
	err = build.writeReservedLBAs()
	if err != nil {
		return err
	}

	// write config
	err = build.writeConfig()
	if err != nil {
		return err
	}

	// write kernel
	err = build.writeKernel()
	if err != nil {
		return err
	}

	// write trampoline
	err = build.writeTrampoline()
	if err != nil {
		return err
	}

	// write app
	err = build.writeApp()
	if err != nil {
		return err
	}

	// write files
	err = build.writeFilesystem()
	if err != nil {
		return err
	}

	return nil

}

func (build *builder) writeReservedLBAs() error {

	var err error

	// write MBR
	err = build.writeMBR()
	if err != nil {
		return err
	}

	// write GPT
	err = build.writeGPT()
	if err != nil {
		return err
	}

	return nil

}
