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

// TODO: tag version

import (
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
)

type cmdTag struct {
	*kingpin.CmdClause
	addr  string
	tag   string
	ref   string
	plain bool
}

// New ...
func newTagCmd() *cmdTag {

	return &cmdTag{}

}

// Attach ...
func (cmd *cmdTag) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("tag", shared.Catenate(`The tag command
		maps an alias or 'tag' for a specific app version's id.`))

	clause := cmd.Arg("address", shared.Catenate(`Address within the local
		repository to tag.`))
	clause.Required()
	clause.StringVar(&cmd.addr)

	clause = cmd.Arg("tag", shared.Catenate(`Tag to use.`))
	clause.Required()
	clause.StringVar(&cmd.tag)

	flag := cmd.Flag("ref", shared.Catenate(`Specific version of the app to
		apply the tag to.`))
	flag.StringVar(&cmd.ref)

	cmd.Action(cmd.action)

}

func (cmd *cmdTag) action(ctx *kingpin.ParseContext) error {

	if strings.HasPrefix(cmd.addr, shared.RepoPrefix) {
		cmd.addr = strings.TrimPrefix(cmd.addr, shared.RepoPrefix)
	}

	mgr, err := vml.NewTinyRepo(home.Path(home.Repository))
	if err != nil {
		return err
	}

	summary, err := mgr.Tag(cmd.addr, cmd.ref, cmd.tag)
	if err != nil {
		return err
	}

	Summary(summary, cmd.plain)

	return nil

}
