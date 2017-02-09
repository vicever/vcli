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
	"fmt"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/shared"
)

type cmdCloudDelete struct {
	*kingpin.CmdClause
	arg string
}

// New ...
func newCloudDeleteCmd() *cmdCloudDelete {

	return &cmdCloudDelete{}

}

// Attach ...
func (cmd *cmdCloudDelete) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("delete", shared.Catenate(`Deletes specified infrastructure.`))
	clause := cmd.Arg("infrastructure", shared.Catenate(`Specificies infrastructure for deletion.`))
	clause.Required()
	clause.PreAction(cmd.preaction)
	clause.StringVar(&cmd.arg)

	cmd.Action(cmd.action)
}

func (cmd *cmdCloudDelete) preaction(ctx *kingpin.ParseContext) error {

	return nil
}

func (cmd *cmdCloudDelete) action(ctx *kingpin.ParseContext) error {

	if cmd.arg == "" {
		fmt.Println("Specify an infrastructure file for deletion.")
	}

	delInf(cmd.arg)

	return nil

}
