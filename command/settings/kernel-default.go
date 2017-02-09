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

type cmdKernelDefault struct {
	*kingpin.CmdClause
	arg          string
	argProvided  bool
	argValidated bool
}

// New ...
func newKernelDefaultCmd() *cmdKernelDefault {

	return &cmdKernelDefault{}

}

// Attach ...
func (cmd *cmdKernelDefault) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("default", shared.Catenate(`The default
		kernel setting defines which version of the vorteil kernel to
		use as the default when launching or building a vorteil
		application. If no arguments are supplied the command prints the
		currently stored value to stdout. Otherwise a string argument
		may be provided to overwrite the currently stored value. Valid
		values for the argument include any string returned by the local
		version of the 'kernel list' command (i.e. using -l).`))

	clause := cmd.Arg("new-value", shared.Catenate(`New value to save as the
		default kernel`))

	// hint sensible default values
	clause.HintAction(home.ListLocalKernels)

	clause.PreAction(cmd.preaction)

	clause.StringVar(&cmd.arg)

	cmd.Action(cmd.action)

}

func (cmd *cmdKernelDefault) preaction(ctx *kingpin.ParseContext) error {

	cmd.argProvided = true
	cmd.argValidated = true

	valid := home.ListLocalKernels()

	for _, x := range valid {

		if x == cmd.arg {
			cmd.argValidated = true
			home.GlobalDefaults.Kernel = cmd.arg
			return nil
		}

	}

	return errors.New(shared.Catenate(`kernel version not found locally; try
		the 'download' command first`))

}

func (cmd *cmdKernelDefault) action(ctx *kingpin.ParseContext) error {

	if !cmd.argProvided {

		if home.GlobalDefaults.Kernel == "" {
			fmt.Println("No kernel files found locally. Try 'vcli settings kernel list' to see available options and then 'vcli settings kernel download <VERSION>' to download one.")
		} else {
			fmt.Println(home.GlobalDefaults.Kernel)
		}

	}

	return nil

}
