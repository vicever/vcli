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

// TODO: list apps
// TODO: list versions

import (
	"errors"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
)

type cmdList struct {
	*kingpin.CmdClause
	addr      string
	absolute  bool
	recursive bool
	long      bool
	plain     bool
}

// New ...
func newListCmd() *cmdList {

	return &cmdList{}

}

// Attach ...
func (cmd *cmdList) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("list", shared.Catenate(`The list command
		prints information about the contents of the local vcli
		repository to the screen.`))

	clause := cmd.Arg("address", shared.Catenate(`Node within the local
		repository to perform the list operation on. Directory nodes
		print child nodes similar to the *nix 'ls' command. App nodes
		print information about the various stored versions of the app.
		If no value is provided then the default value will be the root
		node.`))
	clause.StringVar(&cmd.addr)

	flag := cmd.Flag("recursive", shared.Catenate(`List all of the children
		of a directory node, recursively.`))
	flag.Short('r')
	flag.BoolVar(&cmd.recursive)

	flag = cmd.Flag("long", shared.Catenate(`Print more information.`))
	flag.Short('l')
	flag.BoolVar(&cmd.long)

	flag = cmd.Flag("absolute", shared.Catenate(`Any node names listed will
		be the full path to that node.`))
	flag.Short('a')
	flag.BoolVar(&cmd.absolute)

	cmd.Action(cmd.action)

}

func (cmd *cmdList) action(ctx *kingpin.ParseContext) error {

	if strings.HasPrefix(cmd.addr, shared.RepoPrefix) {
		cmd.addr = strings.TrimPrefix(cmd.addr, shared.RepoPrefix)
	}

	mgr, err := vml.NewTinyRepo(home.Path(home.Repository))
	if err != nil {
		return err
	}

	nt, err := mgr.NodeType(cmd.addr)
	if err != nil {
		return err
	}

	switch nt {
	case "dir":
		list, err := mgr.ListDir(cmd.addr)
		if err != nil {
			return err
		}

		cmd.listDir(list)

	case "app":
		list, err := mgr.ListApp(cmd.addr)
		if err != nil {
			return err
		}

		cmd.listApp(list)

	default:
		return errors.New("mistake")

	}

	return nil

}

func (cmd *cmdList) listDir(list vml.DirectoryList) {
	var data [][]string
	data = [][]string{[]string{"TYPE", "NAME"}}
	for _, x := range list {
		data = append(data, []string{
			x.Type,
			x.Name,
		})
	}
	shared.PrettyLeftTable(data)
}

func (cmd *cmdList) listApp(list vml.AppList) {
	var data [][]string
	data = [][]string{[]string{"DATE", "REFERENCE"}}
	for _, x := range list {
		t := time.Unix(x.Uploaded, 0)
		data = append(data, []string{
			t.Format(time.RFC822),
			x.Reference,
		})
	}
	shared.PrettyLeftTable(data)
}
