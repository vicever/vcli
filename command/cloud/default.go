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
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdDefault struct {
	*kingpin.CmdClause
	arg         string
	argProvided bool
}

// New ...
func newDefaultCmd() *cmdDefault {

	return &cmdDefault{}

}

// Attach ...
func (cmd *cmdDefault) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("default", shared.Catenate(`View or set default infrastructure.`))
	clause := cmd.Arg("infrastructure", shared.Catenate(`Specificies which infrastructure's will be set as default.`))
	clause.PreAction(cmd.preaction)
	clause.StringVar(&cmd.arg)

	cmd.PreAction(cmd.preaction)
	cmd.Action(cmd.action)
}

func (cmd *cmdDefault) preaction(ctx *kingpin.ParseContext) error {

	if _, err := os.Stat(home.Path(home.Infrastructures) + "/" + home.GlobalDefaults.Infrastructure); os.IsNotExist(err) {
		home.GlobalDefaults.Infrastructure = ""
	}

	return nil

}

func (cmd *cmdDefault) action(ctx *kingpin.ParseContext) error {

	if cmd.arg == "" {
		if home.GlobalDefaults.Infrastructure == "" {
			fmt.Println("No default infrastructure file curently set.")
		} else {
			fmt.Printf("Default infrastructure is: %v\n", home.GlobalDefaults.Infrastructure)
		}
	} else {
		if _, err := os.Stat(home.Path(home.Infrastructures) + "/" + cmd.arg); os.IsNotExist(err) {
			return errors.New("failed to set new default file. file '" + cmd.arg + "' does not exist")
		} else {
			home.GlobalDefaults.Infrastructure = cmd.arg
			fmt.Printf("New default infrastructure set to: %v\n", cmd.arg)
		}
	}

	return nil

}
