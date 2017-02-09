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
	"bufio"
	"fmt"
	"io/ioutil"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	ed "github.com/sisatech/vcli/command/config/configui"
	"github.com/sisatech/vcli/shared"
	yaml "gopkg.in/yaml.v2"
)

type configUI struct {
	*kingpin.CmdClause
	reader *bufio.Reader
	arg    string
}

func newEditorCommand() *configUI {

	return &configUI{}
}

func (cmd *configUI) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("config-editor", shared.Catenate(`Launch the
		custom vcli configuration file editor to easily create or edit
		vorteil configuration files from within the commandline.`))

	clause := cmd.Arg("filepath", shared.Catenate(`Targets file for editing. If unused, default values will be used.`))
	clause.Default("")
	clause.StringVar(&cmd.arg)

	cmd.Action(cmd.action)
}

func (cmd *configUI) action(ctx *kingpin.ParseContext) error {

	var file *shared.BuildConfig
	file = new(shared.BuildConfig)
	file.App = new(shared.BuildAppConfig)
	file.Network = new(shared.NetworkConfig)
	file.Disk = new(shared.DiskConfig)
	// file.Network.NetworkCards = new([]shared.NetworkCardConfig)
	file.Network.NetworkCards = make([]shared.NetworkCardConfig, 0)

	if cmd.arg == "" {
		// If no filepath argument ...
		file.Name = "default"
		file.Description = "default"
		file.Author = "default"
		file.Version = "default"
		file.AppURL = "default"

		file.App.BinaryArgs = append(file.App.BinaryArgs, "default")
		file.App.SystemEnvs = append(file.App.SystemEnvs, "default")

		file.Disk.FileSystem = "ext2"
		file.Disk.MaxFD = 128
		file.Disk.DiskSize = 256

		file.Network.DNS = append(file.Network.DNS, "8.8.8.8")

		var card shared.NetworkCardConfig
		card.IP = "192.168.0.1"
		card.Mask = "255.255.0.0"
		card.Gateway = "0.0.255.255"
		file.Network.NetworkCards = append(file.Network.NetworkCards, card)
	} else {
		// If filepath argument provided ...
		buf, err := ioutil.ReadFile(cmd.arg)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(buf, file)
		if err != nil {
			return err
		}
	}

	ed.Edit(file, false)

	fmt.Printf("Name is: %v\n", file.Name)

	out, err := yaml.Marshal(file)
	if err != nil {
		return err
	}

	if cmd.arg != "" {
		err = ioutil.WriteFile(cmd.arg, out, 0666)
		if err != nil {
			return err
		}
	} else {
		err = ioutil.WriteFile(file.Name+".vcfg", out, 0666)
		if err != nil {
			return err
		}
	}

	return nil
}
