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
	"fmt"
	"os/user"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdAuthor struct {
	*kingpin.CmdClause
	arg         string
	argProvided bool
}

// New ...
func newAuthorCmd() *cmdAuthor {

	return &cmdAuthor{}

}

// Attach ...
func (cmd *cmdAuthor) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("author", shared.Catenate(`The author
		setting defines a value to use as the default within a config
		file and anywhere else an author might be used. If no arguments
		are supplied the command pritns the currently stored value to
		stdout. Otherwise a string argument may be provided to overwrite
		the currently stored value.`))

	clause := cmd.Arg("new-value", shared.Catenate(`New value to save as the
		default author.`))

	// hint sensible default values
	usr, err := user.Current()
	if err == nil {
		clause.HintOptions(usr.Name)
	}

	clause.PreAction(cmd.preaction)

	clause.StringVar(&cmd.arg)

	cmd.Action(cmd.action)

}

func (cmd *cmdAuthor) preaction(ctx *kingpin.ParseContext) error {

	cmd.argProvided = true
	home.GlobalDefaults.Author = cmd.arg
	return nil

}

func (cmd *cmdAuthor) action(ctx *kingpin.ParseContext) error {

	if !cmd.argProvided {
		fmt.Println(home.GlobalDefaults.Author)
	}

	return nil

}
