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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/command/config/configui"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdNewConfig struct {
	*kingpin.CmdClause
	arg string
}

// New ...
func newConfigCmd() *cmdNewConfig {

	return &cmdNewConfig{}

}

// Attach ...
func (cmd *cmdNewConfig) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("new", shared.Catenate(`Create a brand
		new config file. If no default config files are set, the new
		file will be set as default.`))

	clause := cmd.Arg("name", shared.Catenate(`Specifies the name of the
		new config file.`))
	clause.StringVar(&cmd.arg)
	clause.Required()

	cmd.Action(cmd.action)

}

func (cmd *cmdNewConfig) argsValidation() error {

	// Check/Handle if cmd.arg is a filepath
	if cmd.arg != "" {

		fp := strings.Split(cmd.arg, "/")
		lp := fp[len(fp)-1]

		// Trim final part of the filepath to get the correct directory
		path := strings.TrimSuffix(cmd.arg, "/"+lp)
		if len(fp) > 1 {
			// If directory does not exist, return err
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return err
			}
		}

	}

	cmd.arg = strings.TrimSuffix(cmd.arg, ".vcfg")

	if _, err := os.Stat(cmd.arg + ".vcfg"); !os.IsNotExist(err) {
		if err == nil {
			return errors.New("file '" + cmd.arg + ".vcfg' already exists. Did you mean to run the 'edit' command?")
		}
		return err
	}

	return nil
}

// Action ...
func (cmd *cmdNewConfig) action(ctx *kingpin.ParseContext) error {

	err := sherlock.Try(func() {

		err := cmd.argsValidation()
		sherlock.Check(err)

		// Create new Config file
		var f *shared.BuildConfig
		f = new(shared.BuildConfig)
		f.Name = "My App"
		f.AppURL = "http://www.example.com/"
		f.Description = shared.Catenate(`Lorem ipsum dolor sit amet, consectetur
		 adipiscing elit. Aliquam pulvinar tortor non sem vestibulum
		 cursus.`)
		f.Author = home.GlobalDefaults.Author
		f.Version = "1.0.0"
		f.ReleaseDate = time.Now()
		f.App = new(shared.BuildAppConfig)
		f.Network = new(shared.NetworkConfig)
		f.Disk = new(shared.DiskConfig)
		f.Network.NetworkCards = make([]shared.NetworkCardConfig, 0)
		f.Network.NetworkCards = append(f.Network.NetworkCards, shared.NetworkCardConfig{
			IP: "dhcp",
		})
		f.Network.DNS = append(f.Network.DNS, "8.8.8.8")
		f.Network.DNS = append(f.Network.DNS, "4.4.4.4")

		f.Disk.FileSystem = "ext2"
		f.Disk.MaxFD = 1024
		f.Disk.DiskSize = 128

		f.NTP = new(shared.NTPConfig)
		f.Redirects = new(shared.RedirectConfig)
		f.Redirects.Rules = make([]shared.Redirect, 0)

		f.NTP.Servers = []string{"0.pool.ntp.org", "1.pool.ntp.org", "2.pool.ntp.org", "3.pool.ntp.org"}

		err = configEditor.Edit(f, true)
		if err != nil && err != configEditor.ErrQ {
			sherlock.Throw(err)
		} else if err != configEditor.ErrQ {

			out, err := json.MarshalIndent(f, "", "  ")
			sherlock.Check(err)
			err = ioutil.WriteFile(cmd.arg+".vcfg", out, 0666)
			sherlock.Check(err)

			fmt.Printf("Created new config file '" + cmd.arg + ".vcfg'.\n")
		}

	})

	return err
}
