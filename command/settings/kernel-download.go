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
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/compiler"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdKernelDownload struct {
	*kingpin.CmdClause
	arg string
}

// New ...
func newKernelDownloadCmd() *cmdKernelDownload {

	return &cmdKernelDownload{}

}

// Attach ...
func (cmd *cmdKernelDownload) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("download", shared.Catenate(`The download
		kernel command attempts to retrieve production and debug
		versions for the specified vorteil kernel version from the
		official Sisa-Tech online repository, so that they can be used
		to build vorteil applications locally.`))

	clause := cmd.Arg("version", shared.Catenate(`The Vorteil kernel version
		to download.`))

	// hint sensible default values
	clause.HintOptions(home.GlobalDefaults.KnownKernelVersions...)

	clause.StringVar(&cmd.arg)

	cmd.Action(cmd.action)

}

func (cmd *cmdKernelDownload) action(ctx *kingpin.ParseContext) error {

	// validation
	if tmp := strings.Split(cmd.arg, "."); len(tmp) != 3 {
		return errors.New("invalid kernel argument")
	} else {
		for _, x := range tmp {
			if _, err := strconv.Atoi(x); err != nil {
				return errors.New("invalid kernel argument")
			}
		}
	}

	// TODO: robust filesystem verification

	// retrieve vtramp if not already found
	_, err := os.Stat(home.Path(home.Kernel + "/vtramp.img"))
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
	prod := "vkernel-PROD-" + cmd.arg + ".img"
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
	debug := "vkernel-DEBUG-" + cmd.arg + ".img"
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

	// set global default to the new kernel if it was previously blank
	if home.GlobalDefaults.Kernel == "" {
		home.GlobalDefaults.Kernel = cmd.arg
	}

	return nil

}
