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
package cmdcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"

	yaml "gopkg.in/yaml.v2"
)

type Infrastructure struct {
	VCenterIP    string `yaml:"vcenter"`
	DataCenter   string `yaml:"datacenter"`
	HostCluster  string `yaml:"hostcluster"`
	Storage      string `yaml:"storagecluster"`
	ResourcePool string `yaml:"resourcepool"`
}

func newInf(name string) {

	sherlock.Try(func() {

		// Init InfraStruct object
		var infraStruct Infrastructure

		// Write a new infrastructure to file ...
		f := home.Path(home.Infrastructures) + "/" + name

		infraStruct.VCenterIP = "default"
		infraStruct.DataCenter = "default"
		infraStruct.HostCluster = "default"
		infraStruct.Storage = "default"
		infraStruct.ResourcePool = "default"

		out, err := json.Marshal(infraStruct)
		sherlock.Check(err)
		err = ioutil.WriteFile(f, out, 0666)
		sherlock.Check(err)
	})

}

func listInf() {
	files, _ := ioutil.ReadDir(home.Path(home.Infrastructures))
	for _, f := range files {
		fmt.Println(f.Name())
	}
}

func delInf(name string) {

	if _, err := os.Stat(home.Path(home.Infrastructures) + "/" + name); os.IsNotExist(err) {
		fmt.Println("target file does not exist")
	} else {
		fp := home.Path(home.Infrastructures) + "/" + name
		if home.GlobalDefaults.Infrastructure == name {
			home.GlobalDefaults.Infrastructure = ""
			fmt.Println("Deleted infrastructure file was set as default. Default infrastructure no longer set. \nYou may set a default infrastructure file by running the 'vcli cloud infrastructure default' command.")
		}

		os.Remove(fp)
	}
}

func infInfo(name string) error {

	err := sherlock.Try(func() {

		var buf []byte
		var err error

		if name == "" {

			if home.GlobalDefaults.Infrastructure == "" {
				sherlock.Throw(errors.New("no default infrastructure file is set"))
			}

			name = home.GlobalDefaults.Infrastructure

			buf, err = ioutil.ReadFile(home.Path(home.Infrastructures) + "/" + name)
			sherlock.Check(err)
		} else {
			buf, err = ioutil.ReadFile(home.Path(home.Infrastructures) + "/" + name)
			sherlock.Check(err)
		}

		typ, err := checkInfType(name)
		sherlock.Check(err)

		if typ == shared.VMWareInf {
			info := new(Infrastructure)
			err = yaml.Unmarshal(buf, info)
			sherlock.Check(err)

			var vals [][]string
			vals = append(vals, []string{"Field", "Value"})
			vals = append(vals, []string{"vCenter", info.VCenterIP})
			vals = append(vals, []string{"datacenter", info.DataCenter})
			vals = append(vals, []string{"cluster", info.HostCluster})
			vals = append(vals, []string{"datastore", info.Storage})
			vals = append(vals, []string{"resource pool", info.ResourcePool})
			shared.PrettyTable(vals)
		} else if typ == shared.GCPInf {
			inf := new(shared.GCPInfrastructure)
			err = yaml.Unmarshal(buf, inf)
			sherlock.Check(err)

			var vals [][]string
			vals = append(vals, []string{"Field", "Value"})
			vals = append(vals, []string{"Bucket", inf.Bucket})
			vals = append(vals, []string{"Zone", inf.Zone})
			shared.PrettyTable(vals)
		}

	})

	return err
}
