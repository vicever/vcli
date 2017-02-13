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
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/vcli/automation"
	"github.com/sisatech/vcli/command/build"
	"github.com/sisatech/vcli/command/cloud"
	"github.com/sisatech/vcli/command/config"
	"github.com/sisatech/vcli/command/repository"
	"github.com/sisatech/vcli/command/run"
	"github.com/sisatech/vcli/command/settings"
	"github.com/sisatech/vcli/home"
)

func main() {

	defer func() {

		r := recover()

		if r != nil {
			fmt.Fprintf(os.Stderr, "an unexpected error occurred: %v\n", r)
		}

	}()

	// initialize environment and packages
	if err := initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	defer cleanup()

	// setup application boilerplate
	app := kingpin.New(vcliName, vcliDesc)
	app.Version(build.Version)
	app.Author(vcliAuthor)

	// change --help output template to use an unflattened command list
	app.UsageTemplate(unflattenedUsageTemplate)

	// registers commands
	cmdrun.New().Attach(app)
	cmdbuild.New().Attach(app)
	cmdvcfg.New().Attach(app)
	cmdrepo.New().Attach(app)
	cmdcloud.New().Attach(app)
	cmdsettings.New().Attach(app)

	// set vmware path for mac
	if runtime.GOOS == "darwin" {
		binPath := "/Applications/VMware Fusion.app/Contents/Library"
		env := os.Getenv("PATH")
		env += ":" + binPath
		os.Setenv("PATH", env)
	}

	// run
	_, err := app.Parse(os.Args[1:])
	if err != nil {

		if os.Args[len(os.Args)-1] == "--completion-bash" || os.Args[len(os.Args)-1] == "--" {
			return
		}

		fmt.Fprintf(os.Stderr, "%v\n", err.Error())
	}

}

func initialize() error {

	var err error
	err = home.Initialize()

	if home.SafeMode {
		return nil
	}

	return err

}

func cleanup() {

	home.Cleanup()

}

// Our own GOVC without forking

// vm-config | vcfg
// 	wizard
// 	templates
// 		import
// 		export
// 		delete
// 		list
//	[saved template]
// repository | repo
// 	import
// 	export
// 	list
// 	delete
// 	tag
// 	untag
// 	filesystem | files | fs
// 		import
// 		export
// 		list
// 		delete
// 	defaults
// 		kernel
// 		vm
// 			cpus
// 			ram
// 			port-mappings
