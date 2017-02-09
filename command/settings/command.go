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

// Command ...
type Command struct {
	*kingpin.CmdClause
}

// New ...
func New() *Command {

	return &Command{}

}

// Attach ...
func (cmd *Command) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("settings", shared.Catenate(`The settings
		command contains subcommands for altering the behaviour of the
		vcli client.`))

	cmd.PreAction(cmd.preaction)

	newAuthorCmd().Attach(cmd)
	newHypervisorsCommand().Attach(cmd)
	newKernelCommand().Attach(cmd)

}

func (cmd *Command) preaction(ctx *kingpin.ParseContext) error {

	return command.NodeOnlyCheck(cmd, ctx)

}
