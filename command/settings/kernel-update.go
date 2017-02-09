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
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/compiler"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdKernelUpdate struct {
	*kingpin.CmdClause
}

// New ...
func newKernelUpdateCmd() *cmdKernelUpdate {

	return &cmdKernelUpdate{}

}

// Attach ...
func (cmd *cmdKernelUpdate) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("update", shared.Catenate(`The update
		kernel command attempts to retrieve production and debug
		versions for the latest vorteil kernel from the official
		Sisa-Tech online repository, so that they can be used to build
		vorteil applications locally. The default kernel setting will
		automatically be changed to match the latest version.`))

	cmd.Action(cmd.action)

}

func (cmd *cmdKernelUpdate) action(ctx *kingpin.ParseContext) error {

	fmt.Printf("Searching for new kernels...\n")

	// get list
	ret, err := compiler.ListKernels()
	if err != nil {
		return err
	}

	version := ret[0].Name

	// download

	// TODO: robust filesystem verification
	fmt.Printf("Downloading Vorteil kernel: %v...\n", version)

	// retrieve vtramp if not already found
	_, err = os.Stat(home.Path(home.Kernel + "/vtramp.img"))
	if err != nil {

		if !os.IsNotExist(err) {
			return errors.New("unexpected error: " + err.Error())
		}

		err = compiler.DownloadVorteilFile("vtramp.img", "vtramp")
		if err != nil {
			return errors.New("unexpected error: " + err.Error())
		}

	}

	// retrieve vboot if not already found
	_, err = os.Stat(home.Path(home.Kernel + "/vboot.img"))
	if err != nil {

		if !os.IsNotExist(err) {
			return errors.New("unexpected error: " + err.Error())
		}

		err = compiler.DownloadVorteilFile("vboot.img", "vboot")
		if err != nil {
			return errors.New("unexpected error: " + err.Error())
		}

	}

	// retrieve prod kernel
	prod := "vkernel-PROD-" + version + ".img"
	_, err = os.Stat(home.Path(home.Kernel + "/" + prod))
	if err != nil {

		if !os.IsNotExist(err) {
			return errors.New("unexpected error: " + err.Error())
		}

		err = compiler.DownloadVorteilFile(prod, "vkernel")
		if err != nil {
			return errors.New("unexpected error: " + err.Error())
		}

	}

	// retrieve debug kernel
	debug := "vkernel-DEBUG-" + version + ".img"
	_, err = os.Stat(home.Path(home.Kernel + "/" + debug))
	if err != nil {

		if !os.IsNotExist(err) {
			return errors.New("unexpected error: " + err.Error())
		}

		err = compiler.DownloadVorteilFile(debug, "vkernel")
		if err != nil {
			return errors.New("unexpected error: " + err.Error())
		}

	}

	// set global default to the new kernel
	home.GlobalDefaults.Kernel = version

	return nil

}
