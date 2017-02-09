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
package home

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"

	"github.com/sisatech/vcli/shared"
)

const (
	globalDefaultsFile = "global.yaml"
)

type globalDefaults struct {
	Author              string   `yaml:"author"`
	Hypervisor          string   `yaml:"hypervisor"`
	Kernel              string   `yaml:"kernel"`
	KnownKernelVersions []string `yaml:"known_kernel_versions"`
	Infrastructure      string   `yaml:"infrastructure"`
}

// GlobalDefaults contains all global fallback values to use when an alternative
// cannot be determined from user input or a more specific saved setting.
var GlobalDefaults globalDefaults

func initGlobalDefaults() error {

	// Init Infrastructure as empty string
	GlobalDefaults.Infrastructure = ""

	info, err := os.Stat(Path(globalDefaultsFile))
	if err != nil {

		if !os.IsNotExist(err) {
			return errors.New("unexpected error: " + err.Error())
		}

	} else if info.IsDir() {

		return errors.New("vcli home directory is corrupt: " + Path(globalDefaultsFile) + " should not be a directory")

	}

	x, err := ioutil.ReadFile(Path(globalDefaultsFile))
	if err != nil {

		// fill with sensible default values
		usr, err := user.Current()
		if err == nil {
			GlobalDefaults.Author = usr.Name
		}

		GlobalDefaults.Kernel = ""

		// choose a default hypervisor
		hypervisors := shared.ListDetectedHypervisors()
		if len(hypervisors) != 0 {
			GlobalDefaults.Hypervisor = hypervisors[0]
		}

		// TODO: check folder / file exists on launch
		// set default infrastructure
		GlobalDefaults.Infrastructure = ""

		return nil

	}

	// parse yaml file
	err = yaml.Unmarshal(x, &GlobalDefaults)
	if err == nil {
		return nil
	}

	// check loaded values for kernel version
	if !ValidLocalKernel(GlobalDefaults.Kernel) {

		vals := ListLocalKernels()
		tmp := ""
		if len(vals) > 0 {
			tmp = vals[0]
		}

		fmt.Fprintf(os.Stdout, "default kernel %v no longer valid; changing default to %v", GlobalDefaults.Kernel, tmp)

	} else if GlobalDefaults.Kernel == "" {

		vals := ListLocalKernels()
		if len(vals) > 0 {
			GlobalDefaults.Kernel = vals[0]
		}

	}

	// check if hypervisor default is still valid
	if !shared.ValidHypervisor(GlobalDefaults.Hypervisor) {

		vals := shared.ListDetectedHypervisors()
		tmp := ""
		if len(vals) > 0 {
			tmp = vals[0]
		}

		fmt.Fprintf(os.Stdout, "default hypervisor %v no longer valid; changing default to %v", GlobalDefaults.Hypervisor, tmp)

	} else if GlobalDefaults.Hypervisor == "" {

		vals := shared.ListDetectedHypervisors()
		if len(vals) > 0 {
			GlobalDefaults.Hypervisor = vals[0]
		}

	}

	return errors.New("vcli home directory is corrupt: " + Path(globalDefaultsFile) + " file is invalid yaml")

}

func saveGlobalDefaults() {

	x, err := yaml.Marshal(&GlobalDefaults)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(Path(globalDefaultsFile), x, perm)
	if err != nil {
		panic(err)
	}

}
