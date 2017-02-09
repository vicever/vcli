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

// TODO: untag version

import (
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
)

type cmdUntag struct {
	*kingpin.CmdClause
	addr  string
	tag   string
	ref   string
	plain bool
}

// New ...
func newUntagCmd() *cmdUntag {

	return &cmdUntag{}

}

// Attach ...
func (cmd *cmdUntag) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("untag", shared.Catenate(`The untag
		command unmaps an alias 'tag' from an application version.`))

	clause := cmd.Arg("address", shared.Catenate(`Address within the local
		repository to tag.`))
	clause.Required()
	clause.StringVar(&cmd.addr)

	clause = cmd.Arg("ref", shared.Catenate(`Specific version of the app to
		untag.`))
	clause.Required()
	clause.StringVar(&cmd.ref)

	cmd.Action(cmd.action)

}

func (cmd *cmdUntag) action(ctx *kingpin.ParseContext) error {

	if strings.HasPrefix(cmd.addr, shared.RepoPrefix) {
		cmd.addr = strings.TrimPrefix(cmd.addr, shared.RepoPrefix)
	}

	mgr, err := vml.NewTinyRepo(home.Path(home.Repository))
	if err != nil {
		return err
	}

	summary, err := mgr.Untag(cmd.addr, cmd.ref)
	if err != nil {
		return err
	}

	Summary(summary, cmd.plain)

	return nil

}
