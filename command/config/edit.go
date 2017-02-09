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
package cmdvcfg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/command/config/configui"
	"github.com/sisatech/vcli/command/repository"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
)

type cmdEditConfig struct {
	*kingpin.CmdClause
	arg    string
	addr   string
	suffix bool
}

// New ...
func editConfigCmd() *cmdEditConfig {

	return &cmdEditConfig{}

}

// Attach ...
func (cmd *cmdEditConfig) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("edit", shared.Catenate(`Edit an existing
		config file.`))

	clause := cmd.Arg("name", shared.Catenate(`Specifies the name of the
		new config file.`))
	clause.StringVar(&cmd.arg)
	clause.Required()

	cmd.Action(cmd.action)

}

func (cmd *cmdEditConfig) argsValidation() error {

	// TODO os.Stat the argument

	cmd.arg = strings.TrimSuffix(cmd.arg, ".vcfg")
	cmd.arg += ".vcfg"

	info, err := os.Stat(cmd.arg)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New(cmd.arg + " is a directory")
	}

	// cmd.suffix = strings.HasSuffix(cmd.arg, ".vcfg")
	// if !cmd.suffix {
	// 	cmd.arg = cmd.arg + ".vcfg"
	// }

	return nil
}

// Action ...
func (cmd *cmdEditConfig) action(ctx *kingpin.ParseContext) error {

	err := sherlock.Try(func() {

		if strings.HasPrefix(cmd.arg, shared.RepoPrefix) {

			cmd.addr = strings.TrimPrefix(cmd.arg, shared.RepoPrefix)

			if strings.HasPrefix(cmd.addr, shared.RepoPrefix) {
				cmd.addr = strings.TrimPrefix(cmd.addr, shared.RepoPrefix)
			}

			mgr, err := vml.NewTinyRepo(cmd.addr)
			sherlock.Check(err)
			out, err := mgr.Export(cmd.addr, "")
			sherlock.Check(err)

			defer out.Close()

			// TODO: use output from Export to build output
			cfg := new(shared.BuildConfig)
			check := new(shared.BuildConfig)
			in, err := ioutil.ReadAll(out.Config())
			sherlock.Check(err)

			err = json.Unmarshal(in, cfg)
			sherlock.Check(err)
			err = json.Unmarshal(in, check)
			sherlock.Check(err)

			configEditor.Edit(cfg, false)

			// check if contents have changed
			if reflect.DeepEqual(cfg, check) {
				fmt.Printf("No changes made to build configuration file.\n")
				return
			}

			tmp, err := json.Marshal(cfg)
			sherlock.Check(err)

			hash, summary, err := mgr.Import(cmd.addr, out.App(), bytes.NewReader(tmp), out.Icon(), out.FilesTar())
			sherlock.Check(err)

			cmdrepo.Summary(summary, false)

			fmt.Println("Imported application with reference id: " + hash)

			return

		}

		cmd.suffix = false
		err := cmd.argsValidation()
		sherlock.Check(err)

		// Create new Config file
		var f *shared.BuildConfig
		f = new(shared.BuildConfig)
		f.App = new(shared.BuildAppConfig)
		f.Network = new(shared.NetworkConfig)
		f.Disk = new(shared.DiskConfig)
		f.Network.NetworkCards = make([]shared.NetworkCardConfig, 0)
		f.NTP = new(shared.NTPConfig)
		f.Redirects = new(shared.RedirectConfig)
		f.Redirects.Rules = make([]shared.Redirect, 0)

		buf, err := ioutil.ReadFile(cmd.arg)
		sherlock.Check(err)
		err = json.Unmarshal(buf, f)
		sherlock.Check(err)
		buf, err = ioutil.ReadFile(cmd.arg)
		sherlock.Check(err)
		err = json.Unmarshal(buf, f)
		sherlock.Check(err)

		err = configEditor.Edit(f, false)
		if err != nil && err == configEditor.ErrQ {
			return
		} else if err != nil {
			sherlock.Throw(err)
		}

		out, err := json.MarshalIndent(f, "", "  ")
		sherlock.Check(err)

		err = ioutil.WriteFile(cmd.arg, out, 0666)
		sherlock.Check(err)

	})

	// fmt.Println("Saved changes to " + cmd.arg)
	return err

}
