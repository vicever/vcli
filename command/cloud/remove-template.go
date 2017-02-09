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
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"

	yaml "gopkg.in/yaml.v2"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/editor/infrastructure"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type cmdRemoveTemplate struct {
	*kingpin.CmdClause
	template       string
	infrastructure string
	vcenter        string
	datacenter     string
	datastore      string
	hostcluster    string
	resourcepool   string

	keyData *shared.GoogleKey
	foo     *os.File
	secure  bool

	envPath string
	inf     *shared.GCPInfrastructure

	username string
	password string
}

// New ...
func newRemoveTemplateCmd() *cmdRemoveTemplate {

	return &cmdRemoveTemplate{}

}

// Attach ...
func (cmd *cmdRemoveTemplate) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("remove-template", shared.Catenate(`Removes an existing template from VSphere.`))
	clause := cmd.Arg("template", shared.Catenate(`Name of the template to target for deletion.`))
	clause.Required()
	clause.StringVar(&cmd.template)
	clause = cmd.Arg("infrastructure", shared.Catenate(`New value.`))
	clause.StringVar(&cmd.infrastructure)
	clause.Default(home.GlobalDefaults.Infrastructure)

	secureflag := cmd.Flag("secure", shared.Catenate(`(VMWARE) If present, VCLI will ignore 'insecure connection' warnings.`))
	secureflag.Short('k')
	secureflag.BoolVar(&cmd.secure)

	cmd.Action(cmd.action)

}

func (cmd *cmdRemoveTemplate) preaction(ctx *kingpin.ParseContext) error {

	return nil
}

func (cmd *cmdRemoveTemplate) action(ctx *kingpin.ParseContext) error {

	err := sherlock.Try(func() {

		var inf infrastructureEditor.Infrastructure

		if cmd.infrastructure == home.GlobalDefaults.Infrastructure && home.GlobalDefaults.Infrastructure == "" {
			// User did not specify an existing infrastructure file, and
			// no default file has been set
			fmt.Println("No infrastructure specified and no default infrastructure set. Please specify an infrastructure file.")
			return
		} else {

			// CALL INF TYPE CHECKER
			t, err := cmd.infTypeChecker(home.Path(home.Infrastructures) + "/" + cmd.infrastructure)
			sherlock.Check(err)

			if t == shared.VMWareInf {
				// yaml.Unmarshal the specified file
				f, err := os.Open(home.Path(home.Infrastructures) + "/" + cmd.infrastructure)
				sherlock.Check(err)
				defer f.Close()

				fs, err := f.Stat()
				sherlock.Check(err)

				r := bufio.NewReader(f)
				b := make([]byte, fs.Size())
				_, err = r.Read(b)
				sherlock.Check(err)

				err = yaml.Unmarshal(b, &inf)
				sherlock.Check(err)
				cmd.vcenter = inf.VCenterIP
				cmd.datacenter = inf.DataCenter
				cmd.datastore = inf.Storage
				cmd.hostcluster = inf.HostCluster
				cmd.resourcepool = inf.ResourcePool

				cmd.username, cmd.password = auth()

				// Template Name ...
				cisp := types.OvfCreateImportSpecParams{
					DiskProvisioning:   "",
					EntityName:         cmd.template,
					IpAllocationPolicy: "",
					IpProtocol:         "",
					OvfManagerCommonParams: types.OvfManagerCommonParams{
						DeploymentOption: "",
						Locale:           "US"},
					PropertyMapping: make([]types.KeyValue, 0),
					NetworkMapping:  make([]types.OvfNetworkMapping, 0),
				}

				cmd.removeTemplate(cisp)

			} else if t == shared.GCPInf {

				// NOTE: GoogleKey file is required for this
				// Unmarshal inf

				// Unmarshal Infrastucture ...
				fullPath := home.Path(home.Infrastructures) + "/" + cmd.infrastructure
				cmd.inf = new(shared.GCPInfrastructure)

				buf, err := ioutil.ReadFile(fullPath)
				sherlock.Check(err)
				err = yaml.Unmarshal(buf, cmd.inf)
				sherlock.Check(err)

				// Write inf.Key to a file, and use that tmpfile as google sysenv Path
				// Also unmarshal into keydata
				tmp, err := ioutil.TempFile("", "")
				sherlock.Check(err)
				_, err = tmp.Write(cmd.inf.Key)
				sherlock.Check(err)

				tbuf, err := ioutil.ReadFile(tmp.Name())
				sherlock.Check(err)

				cmd.keyData = new(shared.GoogleKey)
				err = json.Unmarshal(tbuf, cmd.keyData)
				sherlock.Check(err)
				tmp.Close()
				cmd.envPath = tmp.Name()

				err = cmd.gcpVal()
				sherlock.Check(err)

				// Check if image exists on target Bucket
				jwtBuf := cmd.inf.Key
				sherlock.Check(err)
				clientKey, err := google.JWTConfigFromJSON(jwtBuf, `https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/datastore https://www.googleapis.com/auth/compute https://www.googleapis.com/auth/devstorage.full_control `+
					`https://www.googleapis.com/auth/admin.datatransfer https://www.googleapis.com/auth/admin.directory.customer https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/activity `+
					`https://www.googleapis.com/auth/drive https://www.googleapis.com/auth/drive.metadata https://www.googleapis.com/auth/appstate https://www.googleapis.com/auth/cloud_debugger https://www.googleapis.com/auth/monitoring `+
					`https://www.googleapis.com/auth/trace.append https://www.googleapis.com/auth/cloud.useraccounts https://www.googleapis.com/auth/compute https://www.googleapis.com/auth/devstorage.full_control `+
					`https://www.googleapis.com/auth/datastore https://www.googleapis.com/auth/ndev.cloudman https://www.googleapis.com/auth/ndev.clouddns.readwrite `+
					`https://www.googleapis.com/auth/replicapool https://www.googleapis.com/auth/ndev.cloudman https://www.googleapis.com/auth/cloudruntimeconfig`)
				sherlock.Check(err)
				httpClient := clientKey.Client(oauth2.NoContext)
				cpu, err := compute.New(httpClient)
				sherlock.Check(err)

				_, err = cpu.Images.Delete(cmd.keyData.Project_id, cmd.template).Do()
				sherlock.Check(err)
				fmt.Println("Image deleted.")

				_, err = cpu.InstanceTemplates.Delete(cmd.keyData.Project_id, cmd.template).Do()
				sherlock.Check(err)
				fmt.Println("Instance Template deleted.")

			}
		}
	})
	return err

}

func (cmd *cmdRemoveTemplate) gcpVal() error {

	// Unmarshal Google Key ...
	// if cmd.gooKey != "" {
	// 	// Check exists ...
	// 	if _, err := os.Stat(cmd.gooKey); os.IsNotExist(err) {
	// 		return errors.New("google key file not found")
	// 	}
	//
	// 	// Unmarshal ...
	// 	buf, err := ioutil.ReadFile(cmd.gooKey)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	err = json.Unmarshal(buf, &cmd.keyData)
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// } else {
	// 	return errors.New("google-key flag is required when uploading to the google cloud platform.")
	// }

	if cmd.template == "" {
		return errors.New("google cloud platform functionality requires --template flag.")
	} else {
		cmd.template = strings.ToLower(cmd.template)
		runes := strings.Split(cmd.template, "")
		for _, r := range runes {
			alpha := r >= "a" && r <= "z"
			numeric := r >= "0" && r <= "9"
			dash := r == "-"

			if !(alpha || numeric || dash) {
				return errors.New("--name flag contains illegal characters. Google Cloud Platform supports only alpha-numeric characters, or '-'.")
			}

		}
	}

	return nil

}

func (cmd *cmdRemoveTemplate) infTypeChecker(path string) (string, error) {

	fullpath := home.Path(home.Infrastructures) + "/" + cmd.infrastructure

	if _, err := os.Stat(fullpath); os.IsNotExist(err) {
		return "", errors.New("specified infrastructure does not exist")
	}

	inf := new(shared.VMWareInfrastructure)
	buf, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(buf, inf)
	if err != nil {
		return "", err
	}

	fmt.Println("Infrastructure type detected: " + inf.Type)
	return inf.Type, nil

}

func (cmd *cmdRemoveTemplate) removeTemplate(cisp types.OvfCreateImportSpecParams) {

	err := sherlock.Try(func() {

		ctx := context.TODO()
		ourUrl := "https://" + cmd.username + ":" + cmd.password + "@" + cmd.vcenter + "/sdk"

		u, err := url.Parse(ourUrl)
		sherlock.Check(err)

		client, err := govmomi.NewClient(ctx, u, cmd.secure)
		sherlock.Check(err)

		fnd := find.NewFinder(client.Client, true)
		dcl, err := fnd.Datacenter(ctx, cmd.datacenter)
		sherlock.Check(err)

		fnd.SetDatacenter(dcl)

		ds, err := fnd.Datastore(ctx, cmd.datastore)
		sherlock.Check(err)

		dssplit := strings.Split(ds.String(), "/")
		rmpath := "[" + dssplit[len(dssplit)-1] + "] " + cmd.template

		// Create VirtualMachine object from template name
		vm, err := fnd.VirtualMachine(ctx, "/"+cmd.datacenter+"/vm/"+cmd.template)
		sherlock.Check(err)

		// Delete template from repository
		task, err := vm.Destroy(ctx)
		sherlock.Check(err)

		err = task.Wait(ctx)
		sherlock.Check(err)

		// Create FileManager for deleting the folder of marked template
		fm := object.NewFileManager(client.Client)
		taskfm, err := fm.DeleteDatastoreFile(ctx, rmpath, dcl)
		sherlock.Check(err)

		err = taskfm.Wait(ctx)
		sherlock.Check(err)
	})

	if err != nil {
		fmt.Println(err)
	}

}
