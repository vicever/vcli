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
package cmdbuild

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/compiler"
	"github.com/sisatech/vcli/compiler/converter"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
)

// Command ...
type Command struct {
	*kingpin.CmdClause
	binary string
	files  string
	kernel string
	format string
	debug  bool
	output string
	icon   string
}

// New ...
func New() *Command {

	return &Command{}

}

// Attach ...
func (cmd *Command) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("build", shared.Catenate(`The build command
		makes it possible to compile a standalone program binary into a
		functioning Vorteil application.`))

	// TODO: incorporate building from vcli repository

	arg := cmd.Arg("source", shared.Catenate(`Target application to use for
		build.`))
	arg.Required()
	arg.StringVar(&cmd.binary)

	arg = cmd.Arg("destination", shared.Catenate(`Custom location to put
		compiled file. This argument is not required; vcli will choose a
		sensible output filename if none are provided.`))
	arg.PreAction(cmd.preactionOutput)
	arg.StringVar(&cmd.output)

	flag := cmd.Flag("files", shared.Catenate(`Directory to clone and build
		into teh Vorteil application as the root directory of the
		disk.`))
	flag.ExistingDirVar(&cmd.files)

	// default kernel
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
	flag.Hidden()

	flag = cmd.Flag("format", shared.Catenate(`Specify the output format to
		build for. The default is `+shared.VMDK+`, which outputs a .vmdk
		file. Other options include `+shared.OVA+` (.ova file), and `+shared.GoogleImageFormat+` (Google Cloud Compatible image). You can
		also specify `+shared.ZipArchive+` to put all of the files into
		a zip archive useful for transporting the files and uploading to
		a Vorteil Management System server.`))
	flag.Default(shared.VMDK)
	flag.HintOptions(shared.VMDK, shared.ZipArchive, shared.OVA,
		shared.GoogleImageFormat)
	flag.StringVar(&cmd.format)

	flag = cmd.Flag("icon", shared.Catenate(`Specify a picture file to use
		as the icon for the Vorteil application. This argument is only
		relevant to the zip archive output type. The picture file must
		be a valid .png picture file.`))
	flag.StringVar(&cmd.icon)
	flag.Hidden()

	cmd.Action(cmd.action)

}

func (cmd *Command) preactionOutput(ctx *kingpin.ParseContext) error {

	info, err := os.Stat(cmd.output)
	if err != nil {

		if os.IsNotExist(err) {
			return nil
		}

		return err

	}

	if info.IsDir() {
		return errors.New("cannot overwrite a directory")
	}

	return nil

}

func (cmd *Command) firstTimeKernel() error {

	if cmd.format == shared.ZipArchive {
		return nil
	}

	if home.GlobalDefaults.Kernel == "" {
		fmt.Println("Performing first time setup.")
		fmt.Println("Downloading latest kernel files.")
	}

	// check if kernel exists
	var kernelFound bool
	list := home.ListLocalKernels()
	for _, x := range list {
		if x == cmd.kernel {
			kernelFound = true
		}
	}

	if kernelFound {
		return nil
	}

	ret, err := compiler.ListKernels()
	if err != nil {
		return err
	}

	if len(ret) < 1 {
		return errors.New("remote kernel files not found")
	}

	if home.GlobalDefaults.Kernel == "" {
		cmd.kernel = ret[0].Name
	}

	var index int
	for i, x := range ret {
		if cmd.kernel == x.Name {
			kernelFound = true
			index = i
		}
	}

	if !kernelFound {
		return fmt.Errorf("kernel %v not found locally or remotely", cmd.kernel)
	}

	fmt.Printf("Vorteil Kernel [%v] - %v\n", ret[index].Name, ret[index].Created)
	arg := ret[index].Name

	// TODO: robust filesystem verification

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
	prod := "vkernel-PROD-" + arg + ".img"
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
	debug := "vkernel-DEBUG-" + arg + ".img"
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
		home.GlobalDefaults.Kernel = arg
		cmd.kernel = arg
	}

	return nil

}

func (cmd *Command) validateArgs() error {

	// check config file is up to date ...
	err := shared.VCFGHealthCheck(home.Path(home.Repository), cmd.binary)
	if err != nil {
		return err
	}

	err = cmd.firstTimeKernel()
	if err != nil {
		return err
	}

	if cmd.kernel == "" {
		return errors.New("empty string for kernel version; try 'vcli settings kernel --help' to define a default")
	}

	// files
	if cmd.files != "" {

		fi, err := os.Stat(cmd.files)
		if err != nil {
			return err
		}

		if !fi.IsDir() {
			return errors.New("files flag not a directory")
		}

	}

	return nil

}

func (cmd *Command) action(ctx *kingpin.ParseContext) error {

	var success bool

	defer func() {

		if !success {
			os.Remove(cmd.output)
		}

	}()

	return sherlock.Try(func() {

		config := cmd.binary + ".vcfg"

		var err error
		var output string
		output = cmd.output

		err = cmd.validateArgs()
		if err != nil {
			sherlock.Check(err)
		}

		// load input
		var in converter.Convertible

		if strings.HasPrefix(cmd.binary, shared.RepoPrefix) {

			// load input from repository
			repo, err := vml.NewTinyRepo(home.Path(home.Repository))
			sherlock.Check(err)

			defer repo.Close()

			in, err = repo.Export(strings.TrimPrefix(cmd.binary, shared.RepoPrefix), "")
			sherlock.Check(err)

		} else {

			// load input from filesystem
			if shared.IsELF(cmd.binary) {

				in, err = converter.LoadLoose(cmd.binary, config, cmd.icon, cmd.files)
				if err != nil {
					sherlock.Check(err)
				}

			} else if shared.IsZip(cmd.binary) {

				in, err = converter.LoadZipFile(cmd.binary)
				if err != nil {
					sherlock.Check(err)
				}

			} else {

				sherlock.Check(errors.New("target not a valid ELF or zip archive"))

			}

		}

		defer in.Close()

		switch cmd.format {

		case shared.Loose:

			if output == "" {
				output = shared.NodeName(strings.TrimPrefix(cmd.binary, shared.RepoPrefix))
			}

			err = converter.ExportLoose(in, output)
			if err != nil {
				sherlock.Check(err)
			}

			success = true

		case shared.ZipArchive:

			if output == "" {
				output = shared.NodeName(strings.TrimPrefix(cmd.binary, shared.RepoPrefix)) + ".zip"
			}

			_, err = converter.ExportZipFile(in, output)
			if err != nil {
				sherlock.Check(err)
			}

			success = true

		case shared.VMDK:

			if output == "" {
				output = shared.NodeName(strings.TrimPrefix(cmd.binary, shared.RepoPrefix)) + ".vmdk"
			}

			_, err = converter.ExportSparseVMDK(in, output, cmd.kernel, cmd.debug)
			if err != nil {
				sherlock.Check(err)
			}

			success = true

		case shared.OVA:

			if output == "" {
				output = shared.NodeName(strings.TrimPrefix(cmd.binary, shared.RepoPrefix)) + ".ova"
			}

			_, err = converter.ExportOVA(in, output, cmd.kernel, cmd.debug)
			if err != nil {
				sherlock.Check(err)
			}

			success = true

		case shared.RAWSparse:

			if output == "" {
				output = shared.NodeName(strings.TrimPrefix(cmd.binary, shared.RepoPrefix))
			}

			err := converter.ExportRAWSparse(in, output, cmd.kernel, cmd.debug)
			if err != nil {
				sherlock.Check(err)
			}

			output += ".tar.gz"

			success = true

		// case shared.GoogleImageFormat:
		//
		// 	_, err = compiler.BuildGoogleCloudDisk(cmd.binary, config, cmd.files, cmd.kernel, cmd.debug)
		// 	output = cmd.binary + ".tar.gz"

		default:
			sherlock.Check(errors.New("invalid build format"))
		}

		fmt.Printf("Finished: %v\n", output)

	})

}
