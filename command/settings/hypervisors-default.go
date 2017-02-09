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
package cmdsettings

import (
	"errors"
	"fmt"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdHypervisorsDefault struct {
	*kingpin.CmdClause
	arg          string
	argProvided  bool
	argValidated bool
}

// New ...
func newHypervisorsDefaultCmd() *cmdHypervisorsDefault {

	return &cmdHypervisorsDefault{}

}

// Attach ...
func (cmd *cmdHypervisorsDefault) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("default", shared.Catenate(`The default
		hypervisor setting defines which hypervisor to use as the
		default when launching a vorteil application in a local vm. If
		no arguments are supplied the command prints the currently
		stored value to stdout. Otherwise a string argument may be
		provided to overwrite the currently stored value. Valid values
		for the argument include any string returned by the 'hypervisors
		list' command.`))

	clause := cmd.Arg("new-value", shared.Catenate(`New value to save as the
		default hypervisor.`))

	// hint sensible default values
	clause.HintAction(shared.ListDetectedHypervisors)

	clause.PreAction(cmd.preaction)

	clause.StringVar(&cmd.arg)

	cmd.Action(cmd.action)

}

func (cmd *cmdHypervisorsDefault) preaction(ctx *kingpin.ParseContext) error {

	cmd.argProvided = true
	cmd.argValidated = true

	valid := shared.ListDetectedHypervisorsWithHidden()

	for _, x := range valid {

		if x == cmd.arg {
			cmd.argValidated = true
			home.GlobalDefaults.Hypervisor = cmd.arg
			return nil
		}

	}

	if cmd.arg == shared.KVM || cmd.arg == shared.QEMU ||
		cmd.arg == shared.VirtualBox ||
		cmd.arg == shared.VMwarePlayer ||
		cmd.arg == shared.VMwareWorkstation ||
		cmd.arg == shared.VMwareTest ||
		cmd.arg == shared.KVMClassic {
		return errors.New(cmd.arg + " not detected on PATH")
	}

	return errors.New("bad or unsupported hypervisor")

}

func (cmd *cmdHypervisorsDefault) action(ctx *kingpin.ParseContext) error {

	if !cmd.argProvided {
		fmt.Println(home.GlobalDefaults.Hypervisor)
	}

	return nil

}
