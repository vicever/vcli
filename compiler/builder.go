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
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/sisatech/vcli/shared"
)

type builder struct {
	*Compiler
	buildVariables
	args   *CompileArgs
	config *shared.BuildConfig
	disk   *os.File
	err    error
	output string
}

type buildVariables struct {
	capacity          uint64 // disk capacity in sectors
	seek              uint64 // write head location
	vmdkOverhead      uint64 // reserved disk sectors for vmdk overhead
	appLBAs           lbas
	configLBAs        lbas
	kernelLBAs        lbas
	trampolineLBAs    lbas
	filesystemSectors uint64
	vmdkEntries       uint64
	partition         [2]lbas
	gd                struct {
		start uint64
	}
	gt struct {
		start uint64
	}
	rgd struct {
		start uint64
	}
	rgt struct {
		start uint64
	}
}

type lbas struct {
	first  uint64
	last   uint64
	length uint64
}

func (builder *builder) initialize() {

	err := builder.validateArgs(builder.args)
	if err != nil {
		builder.err = err
		return
	}

	err = builder.loadConfig()
	if err != nil {
		builder.err = err
		return
	}

	// open destination file
	if builder.args.Destination == "" {
		builder.disk, err = ioutil.TempFile("", "compiler-")
	} else {
		builder.disk, err = os.Create(builder.args.Destination)
	}

	if err != nil {
		builder.err = err
		return
	}

	builder.output = builder.disk.Name()

}

func (builder *builder) loadConfig() error {

	builder.config = new(shared.BuildConfig)
	in, err := ioutil.ReadFile(builder.args.Config)
	if err != nil {
		return err
	}

	err = json.Unmarshal(in, builder.config)
	// err = yaml.Unmarshal(in, builder.config)
	if err != nil {
		return err
	}

	return nil

}
