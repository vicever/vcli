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
	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/shared"
)

type cmdHypervisors struct {
	*kingpin.CmdClause
}

func newHypervisorsCommand() *cmdHypervisors {

	return &cmdHypervisors{}

}

// Attach ...
func (cmd *cmdHypervisors) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("hypervisors", shared.Catenate(`The
		hypervisors subcommand contains commands to identify and select
		valid hypervisors usable by vcli, as well as to specify a global
		default hypervisor to use.`))

	cmd.PreAction(cmd.preaction)

	newHypervisorsDefaultCmd().Attach(cmd)
	newHypervisorsListCmd().Attach(cmd)

}

func (cmd *cmdHypervisors) preaction(ctx *kingpin.ParseContext) error {

	return command.NodeOnlyCheck(cmd, ctx)

}
