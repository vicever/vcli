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

// TODO: export file from local repository
// TODO: export support loose files and zip archives alike

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"

	"github.com/sisatech/vcli/command/config/configui"
)

type cmdExport struct {
	*kingpin.CmdClause
	addr string
}

// New ...
func newEditVCFGCmd() *cmdExport {

	return &cmdExport{}

}

// Attach ...
func (cmd *cmdExport) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("edit-config", shared.Catenate(`
		The edit-config command retrieves build configuration file
		for the latest version of the Vorteil app stored in the local
		repository, and runs it through the vcli configuration editor,
		saving any changes as a new version within the repository.`))

	clause := cmd.Arg("address", shared.Catenate(`Address within the local
		repository to lookup the Vorteil app to be exported.`))
	clause.Required()
	clause.StringVar(&cmd.addr)
	// TODO autocomplete

	cmd.Action(cmd.action)

}

func (cmd *cmdExport) action(ctx *kingpin.ParseContext) error {

	if strings.HasPrefix(cmd.addr, shared.RepoPrefix) {
		cmd.addr = strings.TrimPrefix(cmd.addr, shared.RepoPrefix)
	}

	mgr, err := vml.NewTinyRepo(home.Path(home.Repository))
	if err != nil {
		return err
	}

	out, err := mgr.Export(cmd.addr, "")
	if err != nil {
		return err
	}

	defer out.Close()

	// TODO: use output from Export to build output
	cfg := new(shared.BuildConfig)
	check := new(shared.BuildConfig)
	in, err := ioutil.ReadAll(out.Config())
	if err != nil {
		return err
	}

	err = json.Unmarshal(in, cfg)
	if err != nil {
		return err
	}
	err = json.Unmarshal(in, check)
	if err != nil {
		return err
	}

	configEditor.Edit(cfg, false)

	// check if contents have changed
	if reflect.DeepEqual(cfg, check) {
		fmt.Printf("No changes made to build configuration file.\n")
		return nil
	}

	tmp, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	hash, summary, err := mgr.Import(cmd.addr, out.App(), bytes.NewReader(tmp), out.Icon(), out.FilesTar())
	if err != nil {
		return err
	}

	Summary(summary, false)

	fmt.Println("Imported application with reference id: " + hash)

	return nil

}
