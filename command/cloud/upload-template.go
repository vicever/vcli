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
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	compute "google.golang.org/api/compute/v1"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"cloud.google.com/go/storage"

	"github.com/alecthomas/kingpin"
	"github.com/howeyc/gopass"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/compiler"
	"github.com/sisatech/vcli/compiler/converter"
	"github.com/sisatech/vcli/editor/infrastructure"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/progress"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	yaml "gopkg.in/yaml.v2"
)

type cmdUploadTemplate struct {
	*kingpin.CmdClause
	arg          string
	argProvided  bool
	argValidated bool

	name   string
	files  string
	kernel string
	debug  bool
	binary string
	format string
	secure bool
	foo    *os.File
	force  bool

	username string
	password string
	infType  string

	envPath string
	inf     *shared.GCPInfrastructure
	keyData *shared.GoogleKey

	client       *vim25.Client
	Datacenter   *object.Datacenter
	Datastore    *object.Datastore
	resourcepool *object.ResourcePool
}

// New ...
func newUploadTemplateCmd() *cmdUploadTemplate {

	return &cmdUploadTemplate{}

}

var vcenter string
var datastore string
var datacenter string
var hostcluster string
var resourcepool string

// Attach ...
func (cmd *cmdUploadTemplate) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("upload-template", shared.Catenate(`Lorem Ipsum`))

	arg := cmd.Arg("source", shared.Catenate(`Target application to use for
		build.`))
	arg.Required()
	arg.ExistingFileVar(&cmd.binary)

	clause := cmd.Arg("infrastructure", shared.Catenate(`Specifies the infrastructure file to use when building the Vorteil application/template.`))
	clause.StringVar(&cmd.arg)
	clause.Default(home.GlobalDefaults.Infrastructure)
	// clause.PreAction(cmd.preaction)

	nameflag := cmd.Flag("name", shared.Catenate(`Specifies name for new template.`))
	nameflag.StringVar(&cmd.name)
	nameflag.Default("")

	filesflag := cmd.Flag("files", shared.Catenate(`Directory to clone and build
		into the Vorteil application as the root directory of the
		disk.`))
	filesflag.ExistingDirVar(&cmd.files)

	kernelflag := cmd.Flag("kernel", shared.Catenate(`Specify a version of the
		Vorteil kernel to build with instead of using the default. The
		default can be changed using the 'vcli settings kernel default'
		command.`))
	kernelflag.HintAction(home.ListLocalKernels)
	kernelflag.Default(home.GlobalDefaults.Kernel)
	kernelflag.StringVar(&cmd.kernel)

	debugflag := cmd.Flag("debug", shared.Catenate(`Build the Vorteil application
		using the debug version of the kernel instead of the production
		version.`))
	debugflag.Short('d')
	debugflag.BoolVar(&cmd.debug)

	forceflag := cmd.Flag("force", shared.Catenate(`Force VCLI to overwrite existing
		templates/images/files on Google Cloud Platform.`))
	forceflag.Short('f')
	forceflag.BoolVar(&cmd.force)

	secureflag := cmd.Flag("secure", shared.Catenate(`(VMWARE) If present, VCLI will ignore 'insecure connection' warnings.`))
	secureflag.Short('k')
	secureflag.BoolVar(&cmd.secure)

	cmd.Action(cmd.action)

}

func (cmd *cmdUploadTemplate) preaction(ctx *kingpin.ParseContext) error {

	// return errors.New(shared.Catenate(`preaction lorem ipsum`))

	return nil
}

func (cmd *cmdUploadTemplate) action(ctx *kingpin.ParseContext) error {

	err := sherlock.Try(func() {

		var err error
		cmd.infType, err = checkInfType(cmd.arg)
		if err != nil {
			sherlock.Check(err)
		}

		// Build from sourcefile ...
		config := cmd.binary + ".vcfg"

		err = cmd.validateArgs()
		sherlock.Check(err)

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

				in, err = converter.LoadLoose(cmd.binary, config, "", cmd.files)
				sherlock.Check(err)

			} else if shared.IsZip(cmd.binary) {

				in, err = converter.LoadZipFile(cmd.binary)
				sherlock.Check(err)
			} else {

				sherlock.Throw(errors.New("target not a valid ELF or zip archive"))

			}

		}

		defer in.Close()

		if cmd.infType == shared.VMWareInf {
			cmd.foo, err = converter.ExportOVA(in, "", cmd.kernel, cmd.debug)
			sherlock.Check(err)
			defer os.Remove(cmd.foo.Name())
		} else if cmd.infType == shared.GCPInf {
			// Validate input for Google Cloud Platform ...
			err := cmd.gcpValidation()
			sherlock.Check(err)
			err = converter.ExportRAWSparse(in, cmd.binary, cmd.kernel, cmd.debug)
			sherlock.Check(err)
		}

		// Validate Args ...
		cmd.validateArgs()

		if cmd.infType == shared.VMWareInf {

			// function call for vmware logic ...
			cmd.username, cmd.password = auth()
			err := cmd.vmWareLogic()
			sherlock.Check(err)

		} else if cmd.infType == shared.GCPInf {
			// function call for gcp logic ...
			if cmd.name == "" {
				cmd.name = cmd.binary
			}
			err = cmd.gcpLogic()
			sherlock.Check(err)
			os.Remove(cmd.binary + ".tar.gz")
		} else {
			fmt.Println("unknown infrastructure. file may be corrupt.")
		}

	})

	return err

}

func (cmd *cmdUploadTemplate) gcpLogic() error {

	fmt.Println("Infrastructure type detected as Google Cloud Platform.")

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", cmd.envPath)
	// Upload to Google Cloud Storage ...
	ctx := context.Background()

	// Creates a client ...
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Establish connection to the desired bucket ...
	bucket := client.Bucket(cmd.inf.Bucket)
	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Accessing bucket: %s.\n", attrs.Name)

	if !cmd.force {
		q := new(storage.Query)
		objs := bucket.Objects(ctx, q)
		bExists := false

		for i := 0; true; i++ {
			o, err := objs.Next()
			if err != nil {
				break
			}
			s := o.Name
			if err != nil {
				return err
			}
			if s == cmd.binary+".tar.gz" {
				bExists = true
				break
			}
		}
		if bExists {
			return errors.New("file already exists in the bucket. Use the -f flag to force overwrite.")
		}
	}

	fileName := strings.TrimSuffix(cmd.binary, ".tar.gz")
	fileName += ".tar.gz"
	_, fileName = path.Split(fileName)

	fmt.Printf("Filename: %s\n", fileName)

	obj := bucket.Object(fileName)
	obj.Delete(ctx)
	// googleImageName := cmd.name + ".tar.gz"

	r, err := os.Open(fileName)
	if err != nil {
		return err
	}

	// Uploads file to Google Storage bucket ...
	w := obj.NewWriter(ctx)
	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}
	w.Close()

	// Create IMAGE
	// Check if image exists on target Bucket
	jwtBuf := cmd.inf.Key
	clientKey, err := google.JWTConfigFromJSON(jwtBuf, `https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/datastore https://www.googleapis.com/auth/compute https://www.googleapis.com/auth/devstorage.full_control `+
		`https://www.googleapis.com/auth/admin.datatransfer https://www.googleapis.com/auth/admin.directory.customer https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/activity `+
		`https://www.googleapis.com/auth/drive https://www.googleapis.com/auth/drive.metadata https://www.googleapis.com/auth/appstate https://www.googleapis.com/auth/cloud_debugger https://www.googleapis.com/auth/monitoring `+
		`https://www.googleapis.com/auth/trace.append https://www.googleapis.com/auth/cloud.useraccounts https://www.googleapis.com/auth/compute https://www.googleapis.com/auth/devstorage.full_control `+
		`https://www.googleapis.com/auth/datastore https://www.googleapis.com/auth/ndev.cloudman https://www.googleapis.com/auth/ndev.clouddns.readwrite `+
		`https://www.googleapis.com/auth/replicapool https://www.googleapis.com/auth/ndev.cloudman https://www.googleapis.com/auth/cloudruntimeconfig`)
	if err != nil {
		return err
	}
	httpClient := clientKey.Client(oauth2.NoContext)

	fmt.Println("Creating image from uploaded tarball " + fileName + "...")
	cpu, err := compute.New(httpClient)
	if err != nil {
		return err
	}

	_, err = cpu.Images.Get(cmd.keyData.Project_id, cmd.name).Do()
	if err == nil {
		fmt.Println("Image " + cmd.name + " already exists. Deleting...")
		cpu.Images.Delete(cmd.keyData.Project_id, cmd.name).Do()
		for {
			<-time.After(time.Second * 3.0)
			_, err := cpu.Images.Get(cmd.keyData.Project_id, cmd.name).Do()
			if err != nil {
				break
			}
		}
	}

	_, err = cpu.Images.Insert(cmd.keyData.Project_id, &compute.Image{
		Name: cmd.name,
		RawDisk: &compute.ImageRawDisk{
			Source: "https://storage.googleapis.com/" + attrs.Name + "/" + fileName,
		},
	}).Do()
	if err != nil {
		return err
	}

	var imgStatus string
	for imgStatus != "READY" {
		response, err := cpu.Images.Get(cmd.keyData.Project_id, cmd.name).Do()
		if err != nil {
			return err
		}
		imgStatus = response.Status
		<-time.After(time.Second * 3.0)
	}

	// Check Instace Template Exists ...
	_, err = cpu.InstanceTemplates.Get(cmd.keyData.Project_id, cmd.name).Do()
	if err == nil {
		fmt.Println("Instance " + cmd.name + " already exists. Deleting...")
		cpu.InstanceTemplates.Delete(cmd.keyData.Project_id, cmd.name).Do()
		for {
			<-time.After(time.Second * 3.0)
			_, err := cpu.InstanceTemplates.Get(cmd.keyData.Project_id, cmd.name).Do()
			if err != nil {
				break
			}
		}
	}

	// Create New Instance Template ...
	fmt.Println("Creating template from image...")
	insertRequest := cpu.InstanceTemplates.Insert(cmd.keyData.Project_id, &compute.InstanceTemplate{
		Name: cmd.name,
		Properties: &compute.InstanceProperties{
			CanIpForward: false,
			MachineType:  "f1-micro",
			Disks: []*compute.AttachedDisk{&compute.AttachedDisk{
				Type:       "PERSISTENT",
				Boot:       true,
				Mode:       "READ_WRITE",
				AutoDelete: true,
				DeviceName: cmd.name,
				InitializeParams: &compute.AttachedDiskInitializeParams{
					SourceImage: "projects/" + cmd.keyData.Project_id + "/" + cmd.inf.Zone + "/images/" + cmd.name,
					DiskType:    "pd-standard",
					DiskSizeGb:  1,
				},
			}},
			NetworkInterfaces: []*compute.NetworkInterface{
				&compute.NetworkInterface{
					Network:    "projects/" + cmd.keyData.Project_id + "/" + cmd.inf.Zone + "/networks/default",
					Subnetwork: "projects/" + cmd.keyData.Project_id + "/regions/global/subnetworks/default",
					AccessConfigs: []*compute.AccessConfig{
						&compute.AccessConfig{
							Name: "External NAT",
							Type: "ONE_TO_ONE_NAT",
						},
					},
				},
			},
			Scheduling: &compute.Scheduling{
				Preemptible:       false,
				OnHostMaintenance: "MIGRATE",
				AutomaticRestart:  true,
			},
			ServiceAccounts: []*compute.ServiceAccount{
				&compute.ServiceAccount{
					Email: "default",
					Scopes: []string{
						"https://www.googleapis.com/auth/devstorage.read_only",
						"https://www.googleapis.com/auth/logging.write",
						"https://www.googleapis.com/auth/monitoring.write",
						"https://www.googleapis.com/auth/servicecontrol",
						"https://www.googleapis.com/auth/service.management.readonly",
					},
				},
			},
		},
	})
	_, err = insertRequest.Do()
	if err != nil {
		return err
	}

	obj.Delete(ctx)
	fmt.Println("New instance template created: " + cmd.name)

	return nil
}

func (cmd *cmdUploadTemplate) checkGCPExists() error {

	// Unmarshal Infrastucture ...
	fullPath := home.Path(home.Infrastructures) + "/" + cmd.arg
	cmd.inf = new(shared.GCPInfrastructure)

	buf, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(buf, cmd.inf)
	if err != nil {
		return err
	}

	// Write inf.Key to a file, and use that tmpfile as google sysenv Path
	// Also unmarshal into keydata
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	_, err = tmp.Write(cmd.inf.Key)
	if err != nil {
		return err
	}

	tbuf, err := ioutil.ReadFile(tmp.Name())
	if err != nil {
		return err
	}

	cmd.keyData = new(shared.GoogleKey)
	err = json.Unmarshal(tbuf, cmd.keyData)
	if err != nil {
		return err
	}
	tmp.Close()
	cmd.envPath = tmp.Name()

	if !cmd.force {
		// Check if image exists on target Bucket
		jwtBuf := cmd.inf.Key

		clientKey, err := google.JWTConfigFromJSON(jwtBuf, `https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/datastore https://www.googleapis.com/auth/compute https://www.googleapis.com/auth/devstorage.full_control `+
			`https://www.googleapis.com/auth/admin.datatransfer https://www.googleapis.com/auth/admin.directory.customer https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/activity `+
			`https://www.googleapis.com/auth/drive https://www.googleapis.com/auth/drive.metadata https://www.googleapis.com/auth/appstate https://www.googleapis.com/auth/cloud_debugger https://www.googleapis.com/auth/monitoring `+
			`https://www.googleapis.com/auth/trace.append https://www.googleapis.com/auth/cloud.useraccounts https://www.googleapis.com/auth/compute https://www.googleapis.com/auth/devstorage.full_control `+
			`https://www.googleapis.com/auth/datastore https://www.googleapis.com/auth/ndev.cloudman https://www.googleapis.com/auth/ndev.clouddns.readwrite `+
			`https://www.googleapis.com/auth/replicapool https://www.googleapis.com/auth/ndev.cloudman https://www.googleapis.com/auth/cloudruntimeconfig`)
		if err != nil {
			return err
		}
		httpClient := clientKey.Client(oauth2.NoContext)

		cpu, err := compute.New(httpClient)
		if err != nil {
			return err
		}

		fmt.Println("Checking if image or template by this name exists remotely...")
		_, err = cpu.Images.Get(cmd.keyData.Project_id, cmd.name).Do()
		if err == nil {
			return errors.New("Image " + cmd.name + " already exists. Use -f flag to overwrite existing files.")
		}

		_, err = cpu.InstanceTemplates.Get(cmd.keyData.Project_id, cmd.name).Do()
		if err == nil {
			return errors.New("Instance " + cmd.name + " already exists. Use -f flag to overwrite existing files.")
		}
	}
	return nil
}

func (cmd *cmdUploadTemplate) gcpValidation() error {

	if cmd.name == "" {
		_, cmd.name = path.Split(cmd.binary)
	} else {
		cmd.name = strings.ToLower(cmd.name)
		runes := strings.Split(cmd.name, "")
		for _, r := range runes {
			alpha := r >= "a" && r <= "z"
			numeric := r >= "0" && r <= "9"
			dash := r == "-"

			if !(alpha || numeric || dash) {
				return errors.New("--name flag contains illegal characters. Google Cloud Platform supports only alpha-numeric characters, or '-'.")
			}

		}
	}

	err := cmd.checkGCPExists()
	if err != nil {
		return err
	}

	return nil
}

func (cmd *cmdUploadTemplate) vmWareLogic() error {
	// Infrastructure File ...
	var inf infrastructureEditor.Infrastructure
	if cmd.arg == home.GlobalDefaults.Infrastructure && home.GlobalDefaults.Infrastructure == "" {
		// User did not specify an existing infrastructure file
		// and default has not been set
		return errors.New("no infrastructure specified and no default infrastructure set. Please specify an infrastructure file.")
	} else {
		// yaml.Unmarshal the specified file
		f, err := os.Open(home.Path(home.Infrastructures) + "/" + cmd.arg)
		sherlock.Check(err)
		defer f.Close()

		fs, err := f.Stat()
		sherlock.Check(err)

		r := bufio.NewReader(f)
		b := make([]byte, fs.Size())
		_, err = r.Read(b)
		sherlock.Check(err)

		err = yaml.Unmarshal(b, &inf)

		vcenter = inf.VCenterIP
		datastore = inf.Storage
		datacenter = inf.DataCenter
		hostcluster = inf.HostCluster
		resourcepool = inf.ResourcePool
	}

	// Template Name ...
	cisp := types.OvfCreateImportSpecParams{
		DiskProvisioning:   "",
		EntityName:         cmd.name,
		IpAllocationPolicy: "",
		IpProtocol:         "",
		OvfManagerCommonParams: types.OvfManagerCommonParams{
			DeploymentOption: "",
			Locale:           "US"},
		PropertyMapping: make([]types.KeyValue, 0),
		NetworkMapping:  make([]types.OvfNetworkMapping, 0),
	}
	err := cmd.importTemplate(cisp)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *cmdUploadTemplate) importTemplate(cisp types.OvfCreateImportSpecParams) error {

	err := sherlock.Try(func() {

		ctx := context.TODO()
		ourUrl := "https://" + cmd.username + ":" + cmd.password + "@" + vcenter + "/sdk"
		// fmt.Printf("vcenter: %s\n", vcenter)
		u, err := url.Parse(ourUrl)
		sherlock.Check(err)

		client, err := govmomi.NewClient(ctx, u, cmd.secure)
		sherlock.Check(err)

		fnd := find.NewFinder(client.Client, true)
		dcl, err := fnd.Datacenter(ctx, datacenter)
		sherlock.Check(err)

		fnd.SetDatacenter(dcl)

		ds, err := fnd.Datastore(ctx, "/"+datacenter+"/datastore/"+datastore+"/")
		sherlock.Check(err)
		fmt.Printf("Datastore: %v\n", ds)
		hc, err := fnd.HostSystem(ctx, "/"+datacenter+"/host/"+hostcluster+"/")
		sherlock.Check(err)
		fmt.Printf("Host Cluster: %v\n", hc)
		rpl, err := fnd.ResourcePool(ctx, "/"+datacenter+"/host/"+hostcluster+"/Resources/"+resourcepool)
		sherlock.Check(err)
		fmt.Printf("Resource Pool: %v\n", rpl)

		m := object.NewOvfManager(client.Client)

		ovfdescriptorbytes, err := cmd.ReadOvf(cmd.foo.Name())
		sherlock.Check(err)

		spec, err := m.CreateImportSpec(ctx, string(ovfdescriptorbytes), rpl, ds, cisp)
		sherlock.Check(err)

		if spec.Error != nil {
			fmt.Printf("%v\n", errors.New(spec.Error[0].LocalizedMessage))
		}
		if spec.Warning != nil {
			for _, w := range spec.Warning {
				fmt.Printf("Warning: %s\n", w.LocalizedMessage)
			}
		}

		folder, err := fnd.FolderOrDefault(ctx, "")
		sherlock.Check(err)

		lease, err := rpl.ImportVApp(ctx, spec.ImportSpec, folder, hc)
		sherlock.Check(err)

		info, err := lease.Wait(ctx)
		sherlock.Check(err)

		var items []ovfFileItem
		for _, device := range info.DeviceUrl {
			for _, item := range spec.FileItem {
				if device.ImportKey != item.DeviceId {
					continue
				}

				u, err := client.Client.ParseURL(device.Url)
				sherlock.Check(err)

				i := ovfFileItem{
					url:  u,
					item: item,
					ch:   make(chan progress.Report),
				}

				items = append(items, i)
			}
		}

		x := newLeaseUpdater(client.Client, lease, items)
		defer x.Done()

		for _, i := range items {
			err = cmd.Upload(lease, i, client)
			sherlock.Check(err)
		}

		lease.HttpNfcLeaseComplete(ctx)
		vm, err := fnd.VirtualMachine(ctx, "/"+datacenter+"/vm/"+cmd.name)
		sherlock.Check(err)

		err = vm.MarkAsTemplate(ctx)
		sherlock.Check(err)

	})

	return err
}

func auth() (string, string) {

	// construct client
	// request username and password or find them in environment variables
	username := os.Getenv("VORTEIL_USERNAME")
	password := os.Getenv("VORTEIL_PASSWORD")
	if username == "" {
		for {
			fmt.Printf("username: ")
			n, err := fmt.Scanln(&username)
			if err != nil || n != 1 {
				continue
			}
			username = strings.Trim(username, "\n")
			fmt.Printf("password: ")
			pwd, err := gopass.GetPasswd()
			fmt.Println("")
			if err != nil {
				fmt.Println(err)
				continue
			}
			password = string(pwd)
			break
		}
	}

	return username, password
}

func (cmd *cmdUploadTemplate) Upload(lease *object.HttpNfcLease, ofi ovfFileItem, client *govmomi.Client) error {
	item := ofi.item
	file := item.Path

	// Read VMDK
	f, size, _ := cmd.ReadVMDK(file, cmd.foo)
	buf := bytes.NewReader(f)

	opts := soap.Upload{
		ContentLength: size,
	}

	// Non-disk files (such as .iso) use the PUT method.
	// Overwrite: t header is also required in this case (ovftool does the same)
	if item.Create {
		opts.Method = "PUT"
		opts.Headers = map[string]string{
			"Overwrite": "t",
		}
	} else {
		opts.Method = "POST"
		opts.Type = "application/x-vnd.vmware-streamVmdk"
	}

	fmt.Println("Upload complete.")
	return client.Client.Upload(buf, ofi.url, &opts)
}

func (cmd *cmdUploadTemplate) validateArgs() error {

	// check config file is up to date ...
	err := shared.VCFGHealthCheck(home.Path(home.Repository), cmd.binary)
	if err != nil {
		return err
	}

	err = sherlock.Try(func() {

		if _, err := os.Stat(home.Path(home.Infrastructures) + "/" + cmd.arg); os.IsNotExist(err) {
			sherlock.Throw(errors.New("invalid infrastructure argument"))
		}

		err := cmd.firstTimeKernel()
		sherlock.Check(err)

		if cmd.kernel == "" {
			sherlock.Throw(errors.New("empty string for kernel version; try 'vcli settings kernel --help' to define a default"))
		}

	})

	return err
}

func (cmd *cmdUploadTemplate) firstTimeKernel() error {

	if home.GlobalDefaults.Kernel != "" {
		return nil
	}

	fmt.Println("Performing first time setup.")
	fmt.Println("Downloading latest kernel files.")

	ret, err := compiler.ListKernels()
	if err != nil {
		return err
	}

	if len(ret) < 1 {
		return errors.New("remote kernel files not found")
	}

	fmt.Printf("Vorteil Kernel [%v] - %v\n", ret[0].Name, ret[0].Created)
	arg := ret[0].Name

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

func fail(err error) {
	if err != nil {
		_, _, x, _ := runtime.Caller(1)
		fmt.Printf("Error on line: %v\n%v\n", x, err)
		os.Exit(1)
	}
}

func (cmd *cmdUploadTemplate) ReadOvf(fpath string) ([]byte, error) {

	var b []byte

	err := sherlock.Try(func() {

		f, err := os.Open(fpath)
		sherlock.Check(err)
		r := tar.NewReader(f)
		h, err := r.Next()
		sherlock.Check(err)
		b = make([]byte, h.Size)
		_, err = r.Read(b)
		sherlock.Check(err)
		defer f.Close()
	})

	return b, err

}

func (cmd *cmdUploadTemplate) ReadVMDK(fpath string, file *os.File) ([]byte, int64, error) {

	var b []byte
	var h *tar.Header

	err := sherlock.Try(func() {

		f, err := os.Open(cmd.foo.Name())
		sherlock.Check(err)
		r := tar.NewReader(f)
		h, err = r.Next()
		sherlock.Check(err)
		h, err = r.Next()
		sherlock.Check(err)
		b = make([]byte, h.Size)
		_, err = r.Read(b)
		sherlock.Check(err)
		defer f.Close()

	})

	return b, h.Size, err

}

type ovfFileItem struct {
	url  *url.URL
	item types.OvfFileItem
	ch   chan progress.Report
}

func (o ovfFileItem) Sink() chan<- progress.Report {
	return o.ch
}

type leaseUpdater struct {
	client *vim25.Client
	lease  *object.HttpNfcLease

	pos   int64 // Number of bytes
	total int64 // Total number of bytes

	done chan struct{} // When lease updater should stop

	wg sync.WaitGroup // Track when update loop is done
}

func newLeaseUpdater(client *vim25.Client, lease *object.HttpNfcLease, items []ovfFileItem) *leaseUpdater {
	l := leaseUpdater{
		client: client,
		lease:  lease,

		done: make(chan struct{}),
	}

	for _, item := range items {
		l.total += item.item.Size
		go l.waitForProgress(item)
	}

	// Kickstart update loop
	l.wg.Add(1)
	go l.run()

	return &l
}

func (l *leaseUpdater) waitForProgress(item ovfFileItem) {
	var pos, total int64

	total = item.item.Size

	for {
		select {
		case <-l.done:
			return
		case p, ok := <-item.ch:
			// Return in case of error
			if ok && p.Error() != nil {
				return
			}

			if !ok {
				// Last element on the channel, add to total
				atomic.AddInt64(&l.pos, total-pos)
				return
			}

			// Approximate progress in number of bytes
			x := int64(float32(total) * (p.Percentage() / 100.0))
			atomic.AddInt64(&l.pos, x-pos)
			pos = x
		}
	}
}

func (l *leaseUpdater) run() {
	defer l.wg.Done()

	tick := time.NewTicker(2 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-l.done:
			return
		case <-tick.C:
			// From the vim api HttpNfcLeaseProgress(percent) doc, percent ==
			// "Completion status represented as an integer in the 0-100 range."
			// Always report the current value of percent, as it will renew the
			// lease even if the value hasn't changed or is 0.
			percent := int32(float32(100*atomic.LoadInt64(&l.pos)) / float32(l.total))
			err := l.lease.HttpNfcLeaseProgress(context.TODO(), percent)
			if err != nil {
				fmt.Printf("from lease updater: %s\n", err)
			}
		}
	}
}

func (l *leaseUpdater) Done() {
	close(l.done)
	l.wg.Wait()
}
