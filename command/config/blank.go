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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/shared"
)

type cmdBlank struct {
	*kingpin.CmdClause
	arg string
}

// New ...
func newBlankCmd() *cmdBlank {

	return &cmdBlank{}

}

// Attach ...
func (cmd *cmdBlank) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("blank", shared.Catenate(`Create an emtpy
		.vcfg file at the target location. An empty .vcfg file is not a
		valid file and must be filled in before it can be used.`))
	cmd.Alias("empty")

	clause := cmd.Arg("destination", shared.Catenate(`An optional argument
		to specify the location and name of the new .vcfg file.	If
		specified it is the user's responsibility to ensure that it is
		named like <binary>.vcfg where <binary> is the name of the
		compiled binary that will be built into a Vorteil application
		with the config file. If not specified it will be in the current
		directory and named 'blank.vcfg'.`))
	clause.StringVar(&cmd.arg)

	cmd.Action(cmd.action)

}

func (cmd *cmdBlank) action(ctx *kingpin.ParseContext) error {

	err := sherlock.Try(func() {

		standard := false

		if cmd.arg == "" {
			standard = true
			cmd.arg = "blank.vcfg"
		} else if !strings.HasSuffix(cmd.arg, ".vcfg") {
			cmd.arg = cmd.arg + ".vcfg"
		}

		config := &shared.BuildConfig{
			App:     &shared.BuildAppConfig{},
			Network: &shared.NetworkConfig{},
			Disk:    &shared.DiskConfig{},
		}

		config.Network.DNS = append(config.Network.DNS, "8.8.8.8")
		config.Network.DNS = append(config.Network.DNS, "4.4.4.4")

		config.Network.NetworkCards = append(config.Network.NetworkCards, shared.NetworkCardConfig{
			IP: "dhcp",
		})

		config.App.BinaryArgs = append(config.App.BinaryArgs, "")
		config.App.SystemEnvs = append(config.App.SystemEnvs, "")

		config.NTP = new(shared.NTPConfig)
		config.NTP.Hostname = ""
		config.NTP.Servers = []string{}

		config.Redirects = new(shared.RedirectConfig)
		rule := new(shared.Redirect)
		config.Redirects.Rules = append(config.Redirects.Rules, *rule)

		out, err := json.MarshalIndent(config, "", "  ")
		sherlock.Check(err)

		if !strings.HasSuffix(cmd.arg, ".vcfg") {
			cmd.arg += ".vcfg"
		}

		err = ioutil.WriteFile(cmd.arg, out, 0666)
		sherlock.Check(err)

		fmt.Printf("Created empty config file '%v'.", cmd.arg)
		if standard {
			fmt.Printf(" Fill in its fields and change its name to '<your app>.vcfg'.\n")
		} else {
			fmt.Printf(" Fill in its fields to get started.\n")
		}

	})
	return err

}
