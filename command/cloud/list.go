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
	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/shared"
)

type cmdCloudList struct {
	*kingpin.CmdClause
	arg         string
	argProvided bool
}

// New ...
func newCloudListCmd() *cmdCloudList {

	return &cmdCloudList{}

}

// Attach ...
func (cmd *cmdCloudList) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("list", shared.Catenate(`List all infrastructures.`))

	cmd.PreAction(cmd.preaction)
	cmd.Action(cmd.action)
}

func (cmd *cmdCloudList) preaction(ctx *kingpin.ParseContext) error {

	cmd.argProvided = true
	return nil

}

func (cmd *cmdCloudList) action(ctx *kingpin.ParseContext) error {

	listInf()

	return nil

}
