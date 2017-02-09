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
package cmdcloud

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/editor/gcpInf"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type cmdNewGoogle struct {
	*kingpin.CmdClause
	arg          string
	argProvided  bool
	argValidated bool
	gkey         string
	bucket       string
	zone         string
}

// New ...
func newNewGoogleCmd() *cmdNewGoogle {

	return &cmdNewGoogle{}

}

// Attach ...
func (cmd *cmdNewGoogle) Attach(parent command.Node) {

	str := "Create a new Google Cloud Platform infrastructure.\n"

	// cmd.CmdClause = parent.Command("vmware", shared.Catenate(`Create a new infrastructure.
	// 	Infrastructure Settings are handled as follows:`))

	cmd.CmdClause = parent.Command("google", "(Google Cloud Platform) "+str)
	// cmd.CmdClause = parent.Command("gcp", "(Google Cloud Platform) Create a new infrastructure for the Google Cloud Platform.")

	gkey := cmd.Arg("key-file", shared.Catenate(`filepath of the Google
		credentials JSON file`))
	gkey.PreAction(cmd.preaction)
	gkey.Required()
	gkey.ExistingFileVar(&cmd.gkey)

	clause := cmd.Arg("name", shared.Catenate(`Name of the infrastructure
		preset being created`))
	clause.PreAction(cmd.preaction)
	clause.StringVar(&cmd.arg)

	if runtime.GOOS == "windows" {
		flag := cmd.Flag("bucket", shared.Catenate(`Name of bucket on Google
			Cloud Storage.`))
		flag.StringVar(&cmd.bucket)

		flag = cmd.Flag("zone", shared.Catenate(`The zone in which the bucket
			is located (ie. us-east1).`))
		flag.StringVar(&cmd.zone)

		cmd.bucket = strings.ToLower(cmd.bucket)
		cmd.zone = strings.ToLower(cmd.zone)
	}

	cmd.Action(cmd.action)

}

func (cmd *cmdNewGoogle) preaction(ctx *kingpin.ParseContext) error {

	return nil

}

func (cmd *cmdNewGoogle) argVal(ctx *kingpin.ParseContext) error {
	if home.GlobalDefaults.Infrastructure != "" && cmd.arg == "" {
		return errors.New("specify name for new infrastructure")
	}

	vp := strings.Split(cmd.arg, "/")
	if len(vp) > 1 {
		return errors.New("invalid argument, can only contain alpha-numeric characters")
	}

	return nil
}

func (cmd *cmdNewGoogle) action(ctx *kingpin.ParseContext) error {

	err := sherlock.Try(func() {

		err := cmd.argVal(ctx)
		sherlock.Check(err)

		fullpath := home.Path(home.Infrastructures) + "/" + cmd.arg
		if _, err := os.Stat(fullpath); !os.IsNotExist(err) && cmd.arg != "" {
			if err == nil {
				sherlock.Throw(errors.New("file by this name already exists. Please specify an alternative name or delete the existing infrastructure file."))
			}
			sherlock.Check(err)
		}

		if cmd.arg == "" {
			fullpath += "default"
		}

		if runtime.GOOS == "windows" {
			// Read in contents of key file as []byte
			// Save value to finalStruct.Key
			buf, err := ioutil.ReadFile(cmd.gkey)
			sherlock.Check(err)

			// finalStruct.Key = buf

			inf := new(shared.GCPInfrastructure)
			inf.Type = "google"
			inf.Bucket = cmd.bucket
			inf.Zone = cmd.zone
			inf.Key = buf

			out, err := yaml.Marshal(inf)
			sherlock.Check(err)

			if cmd.arg == "" {
				err = ioutil.WriteFile(home.Path(home.Infrastructures)+"/default", out, 0666)
				sherlock.Check(err)
				cmd.arg = "default"
			} else {
				err = ioutil.WriteFile(home.Path(home.Infrastructures)+"/"+cmd.arg, out, 0666)
				sherlock.Check(err)
				fmt.Printf("New infrastructure file, '%s', created at: %s\n", cmd.arg, home.Path(home.Infrastructures))
			}
		} else {
			// Command to call editor on new file.
			err = gcpInf.New(fullpath, cmd.gkey, false)

			if err != nil {
				fmt.Println("User terminated the program." + err.Error())
			} else {
				if cmd.arg != "" {
					if home.GlobalDefaults.Infrastructure == "" {
						home.GlobalDefaults.Infrastructure = cmd.arg
						fmt.Println("New infrastructure settings '" + cmd.arg + "' created and set as default.")
					}
				} else {
					if home.GlobalDefaults.Infrastructure == "" {
						home.GlobalDefaults.Infrastructure = "default"
						fmt.Println("New infrastructure settings 'default' created and set as default.")
					}
				}
			}
		}
	})

	return err

}
