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
	"encoding/json"
	"io/ioutil"

	"github.com/sisatech/vcli/shared"
)

type BuildArgs struct {
	Binary      string
	Config      string
	Files       string
	Kernel      string
	Debug       bool
	Destination string
}

func (build *builder) validateArgs() error {

	err := build.loadConfig()
	if err != nil {
		return err
	}

	// TODO

	// TODO ensure disk has space for 'files' without writing to last grain

	return nil

}

func (build *builder) loadConfig() error {

	build.config = new(shared.BuildConfig)
	in, err := ioutil.ReadFile(build.args.Config)
	if err != nil {
		return err
	}

	err = json.Unmarshal(in, build.config)
	// err = yaml.Unmarshal(in, build.config)
	if err != nil {
		return err
	}

	return nil

}
