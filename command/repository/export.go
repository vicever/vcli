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
	"errors"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/compiler/converter"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
)

type cmdEditVCFG struct {
	*kingpin.CmdClause
	dest   string
	addr   string
	ref    string
	format string
	kernel string
	debug  bool
}

// New ...
func newExportCmd() *cmdEditVCFG {

	return &cmdEditVCFG{}

}

// Attach ...
func (cmd *cmdEditVCFG) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("export", shared.Catenate(`The export
		command takes the target Vorteil application from the local
		repository and copies it to the local filesystem.`))

	clause := cmd.Arg("address", shared.Catenate(`Address within the local
		repository to lookup the Vorteil app to be exported.`))
	clause.Required()
	clause.StringVar(&cmd.addr)
	// TODO autocomplete

	clause = cmd.Arg("destination", shared.Catenate(`Output location on the
		local filesystem to put the exported Vorteil app.`))
	clause.StringVar(&cmd.dest)

	flag := cmd.Flag("ref", shared.Catenate(`Specify a specific version of
		the target Vorteil app to export. A tag or an ID can be used.`))
	flag.StringVar(&cmd.ref)
	// TODO autocomplete

	flag = cmd.Flag("format", shared.Catenate(`Specify the output format to
		build for. The default is `+shared.VMDK+`, which outputs a .vmdk
		file. Other options include `+shared.OVA+` (.ova file). You can
		also specify `+shared.ZipArchive+` to put all of the files into
		a zip archive useful for transporting the files and uploading to
		a Vorteil Management System server.`))
	flag.Default(shared.ZipArchive)
	flag.HintOptions(shared.VMDK, shared.ZipArchive, shared.OVA,
		shared.GoogleImageFormat)
	flag.StringVar(&cmd.format)

	flag = cmd.Flag("kernel", shared.Catenate(`Specify a version of the
		Vorteil kernel to build with instead of using the default. The
		default can be changed using the 'vcli settings kernel default'
		command.`))
	flag.HintAction(home.ListLocalKernels)
	flag.Default(home.GlobalDefaults.Kernel)
	flag.StringVar(&cmd.kernel)

	flag = cmd.Flag("debug", shared.Catenate(`Build the Vorteil application
		using the debug version of the kernel instead of the production
		version.`))
	flag.Short('d')
	flag.BoolVar(&cmd.debug)

	cmd.Action(cmd.action)

}

func (cmd *cmdEditVCFG) action(ctx *kingpin.ParseContext) error {

	if strings.HasPrefix(cmd.addr, shared.RepoPrefix) {
		cmd.addr = strings.TrimPrefix(cmd.addr, shared.RepoPrefix)
	}

	mgr, err := vml.NewTinyRepo(home.Path(home.Repository))
	if err != nil {
		return err
	}

	in, err := mgr.Export(cmd.addr, cmd.ref)
	if err != nil {
		return err
	}

	var output string
	output = cmd.dest

	switch cmd.format {

	case shared.Loose:

		if output == "" {
			output = shared.NodeName(cmd.addr)
		}

		err = converter.ExportLoose(in, output)
		if err != nil {
			return err
		}

	case shared.ZipArchive:

		if output == "" {
			output = strings.TrimSuffix(shared.NodeName(cmd.addr), ".zip") + ".zip"
		}

		_, err = converter.ExportZipFile(in, output)
		if err != nil {
			return err
		}

	case shared.VMDK:

		if output == "" {
			output = strings.TrimSuffix(shared.NodeName(cmd.addr), ".vmdk") + ".vmdk"
		}

		_, err = converter.ExportSparseVMDK(in, output, cmd.kernel, cmd.debug)
		if err != nil {
			return err
		}

	case shared.OVA:

		if output == "" {
			output = strings.TrimSuffix(shared.NodeName(cmd.addr), ".ova") + ".ova"
		}

		_, err = converter.ExportOVA(in, output, cmd.kernel, cmd.debug)
		if err != nil {
			return err
		}

	// case shared.GoogleImageFormat:
	//
	// 	_, err = compiler.BuildGoogleCloudDisk(cmd.binary, config, cmd.files, cmd.kernel, cmd.debug)
	// 	output = cmd.binary + ".tar.gz"

	default:
		return errors.New("invalid build format")
	}

	// TODO: use output from Export to build output

	return nil

}
