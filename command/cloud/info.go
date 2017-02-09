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
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/shared"
)

type cmdCloudInfo struct {
	*kingpin.CmdClause
	arg         string
	argProvided bool
}

// New ...
func newCloudInfoCmd() *cmdCloudInfo {

	return &cmdCloudInfo{}

}

// Attach ...
func (cmd *cmdCloudInfo) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("info", shared.Catenate(`Shows information for specified thing.`))
	clause := cmd.Arg("infrastructure", shared.Catenate(`Specificies which infrastructure's info to display.`))
	clause.PreAction(cmd.preaction)
	clause.StringVar(&cmd.arg)

	cmd.PreAction(cmd.preaction)
	cmd.Action(cmd.action)
}

func (cmd *cmdCloudInfo) preaction(ctx *kingpin.ParseContext) error {

	return nil

}

func (cmd *cmdCloudInfo) action(ctx *kingpin.ParseContext) error {

	err := sherlock.Try(func() {

		err := infInfo(cmd.arg)
		sherlock.Check(err)

	})

	return err

}
