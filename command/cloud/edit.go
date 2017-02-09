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

	yaml "gopkg.in/yaml.v2"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/editor/gcpInf"
	ied "github.com/sisatech/vcli/editor/infrastructure"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdEditVMWare struct {
	*kingpin.CmdClause
	arg          string
	argProvided  bool
	argValidated bool
}

// New ...
func newEditVMWareCmd() *cmdEditVMWare {

	return &cmdEditVMWare{}

}

// Attach ...
func (cmd *cmdEditVMWare) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("edit", shared.Catenate(`Create a new infrastructure.`))

	clause := cmd.Arg("name", shared.Catenate(`Name of the infrastructure file to be edited.`))
	clause.StringVar(&cmd.arg)
	clause.Default(home.GlobalDefaults.Infrastructure)

	cmd.Action(cmd.action)

}

func (cmd *cmdEditVMWare) preaction(ctx *kingpin.ParseContext) error {

	return nil

}

func (cmd *cmdEditVMWare) action(ctx *kingpin.ParseContext) error {

	fullpath := home.Path(home.Infrastructures) + "/" + cmd.arg

	infType, err := cmd.InfTypeCheck(fullpath)
	if err != nil {
		return err
	}

	err = sherlock.Try(func() {

		if _, err := os.Stat(home.Path(home.Infrastructures) + "/" + cmd.arg); os.IsNotExist(err) {
			sherlock.Throw(errors.New(shared.Catenate(`target file does not exist
			(for a list of available infrastructure files, use the
				'cloud infrastructure list' command)`)))
		}

		if infType == shared.VMWareInf {
			err = ied.New(fullpath, true)
		} else {
			// No need for gkey filepath
			err = gcpInf.New(fullpath, "", true)
		}

		if err != nil {
			sherlock.Throw(err)
		} else {
			if cmd.arg != "" {
				if home.GlobalDefaults.Infrastructure == "" {
					home.GlobalDefaults.Infrastructure = cmd.arg
					fmt.Println("New infrastructure file 'infs/" + cmd.arg + "' created and set as default.")
				}
			} else {
				if home.GlobalDefaults.Infrastructure == "" {
					home.GlobalDefaults.Infrastructure = cmd.arg
					fmt.Println("New infrastructure file 'infs/default' created and set as default.")
				}
			}
		}
	})

	return err

}

func (cmd *cmdEditVMWare) InfTypeCheck(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New("infrastructure file does not exist")
	}

	inf := new(shared.VMWareInfrastructure)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(buf, inf)
	if err != nil {
		return "", err
	}

	return inf.Type, nil

}
