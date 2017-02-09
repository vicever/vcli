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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	ied "github.com/sisatech/vcli/editor/infrastructure"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdNewVMWare struct {
	*kingpin.CmdClause
	arg          string
	argProvided  bool
	argValidated bool

	vcenter      string
	datacenter   string
	hostcluster  string
	storage      string
	resourcepool string
}

// New ...
func newNewVMWareCmd() *cmdNewVMWare {

	return &cmdNewVMWare{}

}

// Attach ...
func (cmd *cmdNewVMWare) Attach(parent command.Node) {

	str := "Create a new infrastructure.\nInfrastructure settings are passed to vSphere as follows:\n"
	str += "/<datacenter>/host/<hostcluster>/Resources/<resourcepool>"

	// cmd.CmdClause = parent.Command("vmware", shared.Catenate(`Create a new infrastructure.
	// 	Infrastructure Settings are handled as follows:`))

	cmd.CmdClause = parent.Command("vmware", "(vSphere/vCenter) "+str)
	// cmd.CmdClause = parent.Command("gcp", "(Google Cloud Platform) Create a new infrastructure for the Google Cloud Platform.")

	clause := cmd.Arg("name", shared.Catenate(`Name of the infrastructure file being created`))
	clause.PreAction(cmd.preaction)
	clause.StringVar(&cmd.arg)

	if runtime.GOOS == "windows" {
		flag := cmd.Flag("vcenter", shared.Catenate(`IPAddress:Port of the vcenter/vsphere
			server.`))
		flag.StringVar(&cmd.vcenter)

		flag = cmd.Flag("datacenter", shared.Catenate(`Relative URL pointing to
			the datacenter location.`))
		flag.StringVar(&cmd.datacenter)

		flag = cmd.Flag("hostcluster", shared.Catenate(`Relative URL pointing to
			the hostcluster location,`))
		flag.StringVar(&cmd.hostcluster)

		flag = cmd.Flag("storage", shared.Catenate(`Relative URL pointing to the
			datastore within the datastore cluster.`))
		flag.StringVar(&cmd.storage)

		flag = cmd.Flag("resourcepool", shared.Catenate(`Relative URL pointing to
			the resource pool location.`))
		flag.StringVar(&cmd.resourcepool)
	}

	cmd.Action(cmd.action)

}

func (cmd *cmdNewVMWare) preaction(ctx *kingpin.ParseContext) error {

	return nil

}

func (cmd *cmdNewVMWare) argVal(ctx *kingpin.ParseContext) error {
	if home.GlobalDefaults.Infrastructure != "" && cmd.arg == "" {
		return errors.New("specify name for new infrastructure file")
	}

	vp := strings.Split(cmd.arg, "/")
	if len(vp) > 1 {
		return errors.New("invalid argument, can only contain alpha-numeric characters")
	}

	return nil
}

func (cmd *cmdNewVMWare) action(ctx *kingpin.ParseContext) error {

	err := sherlock.Try(func() {

		err := cmd.argVal(ctx)
		sherlock.Check(err)

		fullpath := home.Path(home.Infrastructures) + "/" + cmd.arg
		if _, err := os.Stat(fullpath); !os.IsNotExist(err) && cmd.arg != "" {
			if err == nil {
				sherlock.Throw(errors.New("file by this name already exists. Please specify an alternative name or delete the existing infrastructure file."))
			}
			sherlock.Check(err)
		}

		if cmd.arg == "" {
			fullpath += "default"
		}

		if runtime.GOOS == "windows" {
			// Create new vmware inf object and populate with provided fields,
			// marshal to infrastructure file...
			inf := new(shared.VMWareInfrastructure)
			inf.VCenterIP = cmd.vcenter
			inf.Type = "vmware"
			inf.Storage = cmd.storage
			inf.ResourcePool = cmd.resourcepool
			inf.HostCluster = cmd.hostcluster
			inf.DataCenter = cmd.datacenter

			out, err := yaml.Marshal(inf)
			sherlock.Check(err)

			if cmd.arg == "" {
				err = ioutil.WriteFile(home.Path(home.Infrastructures)+"/default", out, 0666)
				sherlock.Check(err)
				cmd.arg = "default"
			} else {
				err = ioutil.WriteFile(home.Path(home.Infrastructures)+"/"+cmd.arg, out, 0666)
				sherlock.Check(err)
				fmt.Printf("New infrastructure file, '%s', created at: %s\n", cmd.arg, home.Path(home.Infrastructures))
			}
		} else {
			err = ied.New(fullpath, false)

			if err != nil {
				fmt.Println("User terminated the program.")
			} else {
				if cmd.arg != "" {
					if home.GlobalDefaults.Infrastructure == "" {
						home.GlobalDefaults.Infrastructure = cmd.arg
						fmt.Println("New infrastructure settings '" + cmd.arg + "' created and set as default.")
					}
				} else {
					if home.GlobalDefaults.Infrastructure == "" {
						home.GlobalDefaults.Infrastructure = "default"
						fmt.Println("New infrastructure settings 'default' created and set as default.")
					}
				}
			}
		}

	})

	return err

}
