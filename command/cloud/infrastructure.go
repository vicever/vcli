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
	"runtime"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/shared"
)

type cmdInfrastructure struct {
	*kingpin.CmdClause
}

// New ...
func newInfrastructureCmd() *cmdInfrastructure {

	return &cmdInfrastructure{}

}

// Attach ...
func (cmd *cmdInfrastructure) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("infrastructure", shared.Catenate(`Contains a set of sub-commands related to the management of infrastructure files.`))

	cmd.PreAction(cmd.preaction)

	newNewCmd().Attach(cmd)
	newCloudListCmd().Attach(cmd)
	newDefaultCmd().Attach(cmd)
	newCloudInfoCmd().Attach(cmd)
	newCloudDeleteCmd().Attach(cmd)
	if runtime.GOOS != "windows" {
		newEditVMWareCmd().Attach(cmd)
	}

}

func (cmd *cmdInfrastructure) preaction(ctx *kingpin.ParseContext) error {

	return command.NodeOnlyCheck(cmd, ctx)

}

func (cmd *cmdInfrastructure) action(ctx *kingpin.ParseContext) error {

	return nil

}
