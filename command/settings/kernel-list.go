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
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/compiler"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdKernelList struct {
	*kingpin.CmdClause
	local bool
}

// New ...
func newKernelListCmd() *cmdKernelList {

	return &cmdKernelList{}

}

// Attach ...
func (cmd *cmdKernelList) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("list", shared.Catenate(`The list kernels
		command returns a list of all vorteil kernels on the local
		system as well as more available to download from the Sisa-Tech
		official kernel repository.`))

	flag := cmd.Flag("local", shared.Catenate(`Only list kernels
		found locally.`))
	flag.Short('l')
	flag.BoolVar(&cmd.local)

	cmd.Action(cmd.action)

}

func (cmd *cmdKernelList) action(ctx *kingpin.ParseContext) error {

	if cmd.local {

		vals := home.ListLocalKernels()

		if len(vals) == 0 {
			return errors.New("no kernels found locally")
		}

		for _, x := range vals {
			fmt.Println(x)
		}

	} else {

		ret, err := compiler.ListKernels()
		if err != nil {
			return err
		}

		var vals [][]string
		vals = append(vals, []string{"Release Date", "Version",
			"Downloaded"})
		for _, x := range ret {
			var row []string
			row = append(row, x.Created.Format(time.RFC822))
			row = append(row, x.Name)
			if x.Local {
				row = append(row, "✓")
			} else {
				row = append(row, "")
			}
			vals = append(vals, row)
		}
		shared.PrettyTable(vals)

	}

	return nil

}
