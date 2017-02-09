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
	"github.com/sisatech/vcli/shared"
)

type cmdHypervisorsList struct {
	*kingpin.CmdClause
}

// New ...
func newHypervisorsListCmd() *cmdHypervisorsList {

	return &cmdHypervisorsList{}

}

// Attach ...
func (cmd *cmdHypervisorsList) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("list", shared.Catenate(`The list
		hypervisors command returns a list of all supported hypervisors
		detected on the system. Hypervisors are detected by searching
		for their binaries on the PATH. Supported hypervisors include
		`+shared.KVM+` ( `+shared.BinaryQEMU+`), `+shared.QEMU+` (`+
		shared.BinaryQEMU+`), `+shared.VirtualBox+` (`+
		shared.BinaryVirtualBox+`), `+shared.VMwarePlayer+` (`+
		shared.BinaryVMwarePlayer+`), and `+shared.VMwareWorkstation+
		` (`+shared.BinaryVMwareWorkstation+`).`))

	cmd.Action(cmd.action)

}

func (cmd *cmdHypervisorsList) action(ctx *kingpin.ParseContext) error {

	hypervisors := shared.ListDetectedHypervisors()
	if len(hypervisors) == 0 {
		return errors.New("no supported hypervisors detected")
	}

	for _, x := range hypervisors {
		fmt.Println(x)
	}

	return nil

}
