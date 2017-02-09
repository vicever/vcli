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
package infrastructureEditor

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/editor"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type Infrastructure struct {
	VCenterIP    string `yaml:"vcenter" nav:"vCenter"`
	DataCenter   string `yaml:"datacenter" nav:"datacenter"`
	HostCluster  string `yaml:"hostcluster" nav:"host cluster"`
	Storage      string `yaml:"storagecluster" nav:"storage"`
	ResourcePool string `yaml:"resourcepool" nav:"resource pool"`
}

var (
	infra = Infrastructure{
		VCenterIP:    "",
		DataCenter:   "",
		HostCluster:  "",
		Storage:      "",
		ResourcePool: "",
	}
	output = shared.VMWareInfrastructure{
		Type:         "vmware",
		VCenterIP:    "",
		DataCenter:   "",
		HostCluster:  "",
		Storage:      "",
		ResourcePool: "",
	}
)

func New(name string, edit bool) error {

	if edit {
		// unmarshal file and apply to 'infra'
		buf, err := ioutil.ReadFile(name)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(buf, &infra)
		if err != nil {
			return err
		}
	}

	ed, err := editor.New(infra)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer ed.Cleanup()
	ed.Title("VCLI Infrastructure Client")

	// Top Level Callbacks
	ed.DisplayCallback(vcDisplay, 0)
	ed.EditCallback(vcEdit, "eg. 0.0.0.0:9999", infra.VCenterIP, 0)
	ed.DisplayCallback(dcDisplay, 1)
	ed.EditCallback(dcEdit, "", infra.DataCenter, 1)
	ed.DisplayCallback(hcDisplay, 2)
	ed.EditCallback(hcEdit, "", infra.HostCluster, 2)
	ed.DisplayCallback(scDisplay, 3)
	ed.EditCallback(scEdit, "", infra.Storage, 3)
	ed.DisplayCallback(rpDisplay, 4)
	ed.EditCallback(rpEdit, "", infra.ResourcePool, 4)

	err = ed.Run()
	if err == nil {
		var err2 error
		name, err2 = save(name)
		if err2 != nil {
			return err2
		}

	}
	ed.Log("DONE RUNNING")

	return err
}

func save(name string) (string, error) {
	err := sherlock.Try(func() {

		output.Type = "vmware"
		output.VCenterIP = infra.VCenterIP
		output.Storage = infra.Storage
		output.ResourcePool = infra.Storage
		output.HostCluster = infra.HostCluster
		output.DataCenter = infra.DataCenter

		out, err := yaml.Marshal(output)
		sherlock.Check(err)
		if name == "" {
			err = ioutil.WriteFile(home.Path(home.Infrastructures)+"/default", out, 0666)
			sherlock.Check(err)
			name = "default"
		} else {
			err = ioutil.WriteFile(name, out, 0666)
			sherlock.Check(err)
		}
	})

	return name, err
}

func vcDisplay() string {
	return infra.VCenterIP
}

func vcEdit(str string) error {
	infra.VCenterIP = str
	return nil
}

func dcDisplay() string {
	return infra.DataCenter
}

func dcEdit(str string) error {
	infra.DataCenter = str
	return nil
}

func hcDisplay() string {
	return infra.HostCluster
}

func hcEdit(str string) error {
	infra.HostCluster = str
	return nil
}

func scDisplay() string {
	return infra.Storage
}

func scEdit(str string) error {
	infra.Storage = str
	return nil
}

func rpDisplay() string {
	return infra.ResourcePool
}

func rpEdit(str string) error {
	infra.ResourcePool = str
	return nil
}
