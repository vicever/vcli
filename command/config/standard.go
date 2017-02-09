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
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdStandard struct {
	*kingpin.CmdClause
	arg string
}

// New ...
func newStandardCmd() *cmdStandard {

	return &cmdStandard{}

}

// Attach ...
func (cmd *cmdStandard) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("standard", shared.Catenate(`Create a
		basic usable .vcfg file at the target location. The settings
		within the file should be ok for most common use-cases.`))

	clause := cmd.Arg("destination", shared.Catenate(`An optional argument
		to specify the location and name of the new .vcfg file.	If
		specified it is the user's responsibility to ensure that it is
		named like <binary>.vcfg where <binary> is the name of the
		compiled binary that will be built into a Vorteil application
		with the config file. If not specified it will be in the current
		directory and named 'standard.vcfg'.`))
	clause.StringVar(&cmd.arg)

	cmd.Action(cmd.action)

}

func (cmd *cmdStandard) action(ctx *kingpin.ParseContext) error {

	err := sherlock.Try(func() {

		standard := false

		if cmd.arg == "" {
			standard = true
			cmd.arg = "standard.vcfg"
		} else if !strings.HasSuffix(cmd.arg, ".vcfg") {
			cmd.arg = cmd.arg + ".vcfg"
		}

		config := &shared.BuildConfig{
			Name:        "My Vorteil Application",
			Description: "My fantastic little Vorteil Application!",
			Author:      home.GlobalDefaults.Author,
			ReleaseDate: time.Now(),
			Version:     "1.0.0",
			AppURL:      "",
			App: &shared.BuildAppConfig{
				// BinaryType: "go",
				BinaryArgs: []string{"myapp"},
				SystemEnvs: []string{},
			},
			Network: &shared.NetworkConfig{
				DNS: []string{},
				NetworkCards: []shared.NetworkCardConfig{
					shared.NetworkCardConfig{
						IP: "dhcp",
					},
				},
			},
			Disk: &shared.DiskConfig{
				FileSystem: "ext2",
				MaxFD:      1024,
				DiskSize:   128,
			},
			Redirects: &shared.RedirectConfig{},
			NTP:       &shared.NTPConfig{},
		}

		config.Network.DNS = append(config.Network.DNS, "8.8.8.8")
		config.Network.DNS = append(config.Network.DNS, "4.4.4.4")

		out, err := json.MarshalIndent(config, "", "  ")
		sherlock.Check(err)

		err = ioutil.WriteFile(cmd.arg, out, 0666)
		sherlock.Check(err)

		fmt.Printf("Created a basic config file '%v' with simple default values.", cmd.arg)
		if standard {
			fmt.Printf(" Adjust its fields and change its name to '<your app>.vcfg'.\n")
		} else {
			fmt.Printf(" Adjust its fields to get started.\n")
		}

	})

	return err

}
