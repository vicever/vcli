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

type cmdKernel struct {
	*kingpin.CmdClause
}

func newKernelCommand() *cmdKernel {

	return &cmdKernel{}

}

// Attach ...
func (cmd *cmdKernel) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("kernel", shared.Catenate(`The kernel
		subcommand contains commands to commands to manage, download,
		and select different versions of the vorteil kernel to improve
		quality assurance testing.`))

	cmd.PreAction(cmd.preaction)

	newKernelDefaultCmd().Attach(cmd)
	newKernelListCmd().Attach(cmd)
	newKernelDownloadCmd().Attach(cmd)
	newKernelUpdateCmd().Attach(cmd)

}

func (cmd *cmdKernel) preaction(ctx *kingpin.ParseContext) error {

	return command.NodeOnlyCheck(cmd, ctx)

}
