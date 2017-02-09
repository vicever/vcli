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
package cmdrepo

// TODO: delete version
// TODO: delete app
// TODO: delete recursively

import (
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
)

type cmdDelete struct {
	*kingpin.CmdClause
	addr      string
	ref       string
	force     bool
	recursive bool
}

// New ...
func newDeleteCmd() *cmdDelete {

	return &cmdDelete{}

}

// Attach ...
func (cmd *cmdDelete) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("delete", shared.Catenate(`The delete
		command is used to remove elements of vcli's local repository.
		The delete command is permanent and cannot be undone, even with
		vcli's version control. By default only application nodes are
		valid targets, and all versions of the application will be
		deleted. vcli automatically recursively removes directories that
		have no children.`))

	clause := cmd.Arg("address", shared.Catenate(`Address within the local
			repository to store the imported app.`))
	clause.Required()
	clause.StringVar(&cmd.addr)

	flag := cmd.Flag("ref", shared.Catenate(`Specify a specific version to
		be deleted, leaving the rest of the app versions intact.`))
	flag.StringVar(&cmd.ref)

	flag = cmd.Flag("force", shared.Catenate(`Perform all actions without
		prompting the user for any confirmations.`))
	flag.BoolVar(&cmd.force)

	flag = cmd.Flag("recursive", shared.Catenate(`Delete a directory and
		all of its children recursively.`))
	flag.BoolVar(&cmd.recursive)

	cmd.Action(cmd.action)

}

func (cmd *cmdDelete) action(ctx *kingpin.ParseContext) error {
	if strings.HasPrefix(cmd.addr, shared.RepoPrefix) {
		cmd.addr = strings.TrimPrefix(cmd.addr, shared.RepoPrefix)
	}
	mgr, err := vml.NewTinyRepo(home.Path(home.Repository))
	if err != nil {
		return err
	}
	summary, err := mgr.Delete(cmd.addr, cmd.ref)
	if err != nil {
		return err
	}
	Summary(summary, false)
	return nil

}
