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
package cmdrun

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin"

	"github.com/hpcloud/tail"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/compiler"
	"github.com/sisatech/vcli/compiler/converter"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
)

var fullPath string
var vBoxSerial *os.File
var dbf *os.File
var vmwname string

// Command ...
type Command struct {
	*kingpin.CmdClause
	binary     string
	files      string
	hypervisor string
	kernel     string
	debug      bool
	persist    string
	memory     uint32
	cpus       uint16
	pmap       string
	headless   bool
	echo       bool
	tempdir    bool
	config     string
}

// New ...
func New() *Command {

	return &Command{}

}

// Attach ...
func (cmd *Command) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("run", shared.Catenate(`The run command
		launches a Vorteil application on a local hypervisor.`))
	cmd.Alias("start")
	cmd.Alias("launch")

	clause := cmd.Arg("application", shared.Catenate(`Target application to
		launch in a local hypervisor.`))
	clause.Required()
	clause.StringVar(&cmd.binary)

	flag := cmd.Flag("files", shared.Catenate(`Directory to clone for use as
		the root directory when compiling the Vorteil application.`))
	flag.ExistingFileOrDirVar(&cmd.files)

	flag = cmd.Flag("hypervisor", shared.Catenate(`Hypervisor to launch
		Vorteil appliation in. Use 'vcli settings hypervisors list' to
		see what options are available on your system.`))
	flag.Default(home.GlobalDefaults.Hypervisor)
	flag.HintAction(shared.ListDetectedHypervisors)
	flag.StringVar(&cmd.hypervisor)

	flag = cmd.Flag("kernel", shared.Catenate(`Version of the Vorteil kernel
		to use when building the Vorteil application before
		launching.`))
	flag.Default(home.GlobalDefaults.Kernel)
	flag.HintAction(home.ListLocalKernels)
	flag.StringVar(&cmd.kernel)

	flag = cmd.Flag("echo", shared.Catenate(`Echo VM's stderrr and stdout to standard output`))
	flag.Short('e')
	flag.BoolVar(&cmd.echo)

	flag = cmd.Flag("headless", shared.Catenate(`Set Virtual Machine to headless.`))
	flag.Short('D')
	flag.BoolVar(&cmd.headless)

	flag = cmd.Flag("debug", shared.Catenate(`Use the debug version of the
		kernel.`))
	flag.Short('d')
	flag.BoolVar(&cmd.debug)
	flag.Hidden()

	flag = cmd.Flag("persist", shared.Catenate(`Place the virtual disk file
		at the target location and leave it there, rather than cleaning
		it up after the hypervisor is closed.`))
	flag.StringVar(&cmd.persist)

	// TODO: some sort of defaults for the following vm settings flags
	flag = cmd.Flag("ram", shared.Catenate(`RAM in megabytes to assign to
		the VM.`))
	flag.Default("64")
	flag.Uint32Var(&cmd.memory)

	flag = cmd.Flag("cpus", shared.Catenate(`Number of virtual cpus to
		assign to the VM.`))
	flag.Default("1")
	flag.Uint16Var(&cmd.cpus)
	flag.Hidden()

	flag = cmd.Flag("port-map", shared.Catenate(`Must be in the format "x:y:z" where x = Network Card number, y = the Host Port, and x = the Guest Port.
		Mappings should be separated with a ',' (ie. --port-map=x:y:z,a:b:c)'`))
	flag.StringVar(&cmd.pmap)

	flag = cmd.Flag("config", shared.Catenate(`Override the default config
		file associated with the app. Specifies path to desired config.`))
	flag.Short('c')
	flag.StringVar(&cmd.config)

	cmd.Action(cmd.action)

}

func (cmd *Command) firstTimeKernel() error {

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
	if cmd.config == "" {
		cmd.config = cmd.binary + ".vcfg"
	}
	err := shared.VCFGHealthCheck(home.Path(home.Repository), cmd.config, false)
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

	cmd.hypervisor = strings.ToUpper(cmd.hypervisor)

	hypervisors := shared.ListDetectedHypervisorsWithHidden()
	hypcheck := false
	for _, hypervisor := range hypervisors {
		if strings.ToLower(cmd.hypervisor) == strings.ToLower(hypervisor) {
			hypcheck = true
		}
	}
	if !hypcheck {
		binhyp := ""
		switch cmd.hypervisor {
		case shared.KVMClassic:
			binhyp = shared.BinaryQEMU
		case shared.KVM:
			binhyp = shared.BinaryQEMU
		case shared.QEMU:
			binhyp = shared.BinaryQEMU
		case shared.VMwarePlayer:
			binhyp = shared.BinaryVMwarePlayer
		case shared.VMwareWorkstation:
			binhyp = shared.BinaryVMwareWorkstation
		case shared.VMwareTest:
			binhyp = shared.BinaryVMwareWorkstation
		case shared.VirtualBox:
			binhyp = shared.BinaryVirtualBox
		default:
			return fmt.Errorf("invalid hypervisor flag")

		}
		return errors.New("hypervisor '" + binhyp + "' not found in path'")
	}

	// Validate sufficient RAM
	// Must be at least 2MB (for kernel) + 2 MB per CPU + size of app, rounded up to nearest 2
	// ram := cmd.memory
	// if ram < (uint32(appsizeMB) + 2*(uint32(cmd.cpus)) + 2) {
	// 	return errors.New("insufficient RAM allocated to the virtual machine.")
	// }
	if cmd.memory < 16 {
		return errors.New("insufficient RAM. Must be at least 16 MB.")
	}
	if cmd.memory%4 != 0 {
		return errors.New("RAM must be a multiple of 4")
	}

	if (cmd.hypervisor == shared.KVMClassic || cmd.hypervisor == shared.QEMU || cmd.hypervisor == shared.VirtualBox) && cmd.pmap != "" {
		pmap := strings.Split(cmd.pmap, ",")
		if len(pmap) > 0 {
			for i, p := range pmap {
				pmapIndividual := strings.Split(p, ":")
				if len(pmapIndividual) != 3 {
					return errors.New("argument '--port-map' formatted incorrectly. Correct format should be '<cardNumber:hostPort:guestPort>', with additional mappings separated by commas.")
				} else {
					a, err := strconv.Atoi(pmapIndividual[1])
					if err != nil {
						return errors.New("unexpected value in port-map flag. Values should be numeric only.")
					}
					b, err := strconv.Atoi(pmapIndividual[2])
					if err != nil {
						return errors.New("unexpected value in port-map flag. Values should be numeric only.")
					}

					if a < 0 || a > 65535 {
						return errors.New(fmt.Sprintf("port-map %v: host port out of bounds. Must be between 0-65535.", i))
					} else if b < 0 || b > 65535 {
						return errors.New(fmt.Sprintf("port-map %v: guest port out of bounds. Must be between 0-65535.", i))
					}
				}
			}
		}
	} else if (cmd.hypervisor == shared.VMwarePlayer || cmd.hypervisor == shared.VMwareWorkstation) && cmd.pmap != "" {
		fmt.Println("WARNING: flag '--port-map' is ignored while hypervisor is set to VMWARE_PLAYER or VMWARE.")
	}

	kern := strings.Split(cmd.kernel, ".")
	if len(kern) != 3 {
		return errors.New("invalid kernel argument syntax. Argument should conform to the format: 'major.minor.patch' (ie. 0.0.1)")
	}

	// Validate Files (handle no files flag given)
	if cmd.files == "" {
		// Create temp dir to be used instead
		t := os.TempDir()
		f, err := ioutil.TempDir(t, "files")
		if err != nil {
			return err
		}
		cmd.files = f
		cmd.tempdir = true
	} else {
		fi, err := os.Stat(cmd.files)
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			return fmt.Errorf("files flag: '%v' is not a directory", cmd.files)
		}
	}

	if cmd.headless && cmd.hypervisor == shared.VMwarePlayer {
		fmt.Println("WARNING: VMWARE_PLAYER does not support headless mode. Continuing with GUI enabled.")
		cmd.headless = false
	}

	return nil

}

func (cmd *Command) action(ctx *kingpin.ParseContext) error {

	return sherlock.Try(func() {

		err := cmd.validateArgs()
		if err != nil {
			sherlock.Check(err)
		}

		// TODO: support launching any and all of the supported formats
		// TODO: argument validation
		config := cmd.config

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

			if _, err := os.Stat(cmd.binary); os.IsNotExist(err) {
				sherlock.Check(errors.New("target not found"))
			}

			// load input from filesystem
			if shared.IsELF(cmd.binary) {

				_, err = os.Stat(config)
				if os.IsNotExist(err) {
					sherlock.Check(fmt.Errorf("no configuration file at \"%v\", try 'vcli vm-config new %v' to create one", config, config))
				}

				in, err = converter.LoadLoose(cmd.binary, config, "", cmd.files)
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

		vmdkPath, err := converter.ExportSparseVMDK(in, cmd.persist, cmd.kernel, cmd.debug)
		if err != nil {
			sherlock.Check(err)
		}
		vmdkPath.Close()

		name := vmdkPath.Name()

		if cmd.tempdir {
			os.Remove(cmd.files)
		}

		if err != nil {
			sherlock.Check(err)
		}

		// TODO: move file before launch if persist flag set

		// TODO: export into any filetype?

		// TODO: check compile target is elf
		// TODO: check disk size is greater than files size

		if cmd.persist == "" {
			if cmd.hypervisor != shared.VMwarePlayer && cmd.hypervisor != shared.VMwareWorkstation && cmd.hypervisor != shared.VMwareTest {
				defer os.Remove(name)
			}
		}

		err = cmd.start(name)
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}

	})

}

// TODO: cleanup following code
func (cmd *Command) start(disk string) error {

	fmt.Printf("Using VMDK: %s\n", disk)

	// executable := "qemu-system-x86_64"
	executable := "qemu-system-x86_64"
	var args []string

	coresString := strconv.FormatUint(uint64(cmd.cpus), 10)
	memoryString := strconv.FormatUint(uint64(cmd.memory), 10)
	appName, err := compiler.ReadAppNameFromVMDK(disk)
	sherlock.Check(err)
	numberOfNetworkCards, err := compiler.ReadNetworkCardCountFromVMDK(disk)

	sherlock.Check(err)

	// var tcpserver net.Listener

	f, err := os.Open(disk)
	sherlock.Check(err)
	f.Close()

	switch cmd.hypervisor {
	case shared.VMwarePlayer:
		executable = "vmplayer"
		args, err = cmd.paramsVMWare(coresString, memoryString, f.Name(), appName, true, numberOfNetworkCards)
		sherlock.Assert(errors.New("error preparing vmware-player workstation"), err == nil)
	case shared.VirtualBox:
		// TODO: fix up hypervisor call
		if !cmd.headless {
			executable = "VirtualBox"
		} else {
			executable = "VBoxManage"
		}

		args, err = cmd.paramsVirtualBox(coresString, memoryString, f.Name(), appName, numberOfNetworkCards, cmd.pmap)
		if err != nil {
			fmt.Printf("Error preparing VirtualBox workstation: %v\n", err)
			panic(err)
		}
	case shared.KVMClassic:
		args = cmd.paramsKVM(coresString, memoryString, f.Name(), numberOfNetworkCards, cmd.pmap)
		// case hypervisorKVMCLOUD:
		// 	args = paramsKVMCloud(coresString, memoryString, f.Name(), numberOfNetworkCards, *cmd.pmap)
	case shared.KVM:
		args = cmd.paramsKVM(coresString, memoryString, f.Name(), numberOfNetworkCards, cmd.pmap)
	case shared.QEMU:
		var dbs string
		if cmd.debug {
			var err error
			dbf, err = ioutil.TempFile("", "QEMU-")
			sherlock.Check(err)
			dbf.Close()
			dbs = dbf.Name()
			fmt.Println("QEMU debug file: " + dbs)
		}
		args = cmd.paramsQemu(coresString, memoryString, f.Name(), dbs, numberOfNetworkCards, cmd.pmap)

	case shared.VMwareWorkstation:
		executable = "vmrun"
		args, err = cmd.paramsVMWare(coresString, memoryString, f.Name(), appName, false, numberOfNetworkCards)
		sherlock.Assert(errors.New("error preparing vmware-workstation workstation"), err == nil)
	case shared.VMwareTest:
		executable = "vmrun"
		args, err = cmd.paramsVMWare(coresString, memoryString, f.Name(), appName, false, numberOfNetworkCards)
		sherlock.Assert(errors.New("error preparing vmware-workstation workstation"), err == nil)
	default:
		return errors.New("unsupported hypervisor")
	}

	// cleanup vmware and virtual box
	if cmd.hypervisor == shared.VMwarePlayer || cmd.hypervisor == shared.VMwareWorkstation || cmd.hypervisor == shared.VMwareTest {

		// Prevents VCLI from cleaning up VMware Fusion instances after
		// run process ends
		if runtime.GOOS != "darwin" {
			command := exec.Command("vmrun", "stop", args[len(args)-1], "hard")
			defer command.Run()
		}

	} else if cmd.hypervisor == shared.VirtualBox {

		command := exec.Command("vboxmanage", "unregistervm", f.Name(), "--delete")
		defer command.Run()

		command = exec.Command("vboxmanage", "controlvm", f.Name(), "poweroff")
		defer command.Run()

	}

	command := exec.Command(executable, args...)

	if (cmd.hypervisor == shared.VirtualBox) && cmd.headless {
		command = exec.Command("vboxmanage", "startvm", appName, "--type", "headless")
	}

	// returns a pipe that will be connected to the command's standard error
	// when the command starts.
	var rc io.ReadCloser
	if cmd.headless {
		rc, err = command.StderrPipe()
		if err != nil {
			panic(err)
		}
	}

	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	chBool := make(chan bool, 1)

	// launch hypervisor and also listen for interrupts
	// fmt.Printf("Args are: %v\n", args)
	if cmd.headless {
		err = command.Start()
	} else {
		var errb bytes.Buffer

		command.Stderr = &errb

		if (cmd.hypervisor == shared.KVMClassic || cmd.hypervisor == shared.QEMU || cmd.hypervisor == shared.KVM) && cmd.echo {
			stdout, err := command.StdoutPipe()
			if err != nil {
				panic(err)
			}

			go func() {
				io.Copy(os.Stdout, stdout)
			}()

		} else {
			var errS bytes.Buffer
			command.Stdout = &errS
			go cmd.serialOut()
		}
		err = command.Run()

		fmt.Printf("\n%v", command.Stderr)

		go func() {

			command.Wait()

			if cmd.hypervisor == shared.VMwareWorkstation || cmd.hypervisor == shared.VMwareTest {

				for {
					time.Sleep(time.Second)

					running := false
					var out bytes.Buffer
					checklist := exec.Command("vmrun", "list")
					checklist.Stdout = &out
					err := checklist.Run()
					if err != nil {
						fmt.Println(err)
					}

					str := strings.Split(out.String(), "\n")
					for _, ln := range str {
						if ln == vmwname {
							running = true
						}
					}

					if !running {
						break
					}

				}
			}

			chBool <- true

		}()

		select {
		case <-signal_channel:
		case <-chBool:
		}

		if cmd.hypervisor == shared.VirtualBox {
			// fmt.Println(f.Name())
			cmD := exec.Command("VBoxManage", "controlvm", appName, "poweroff")
			cmD.Run()
			cmD = exec.Command("VBoxManage", "unregistervm", appName)
			cmD.Run()

			if cmd.echo {
				os.Remove(vBoxSerial.Name())
			}
		}
		if cmd.debug && cmd.hypervisor == shared.QEMU {
			// Don't delete the QEMU log file after process finishes ...
			if !cmd.debug {
				os.Remove(dbf.Name())
			}
		}

	}
	if err != nil {
		// fmt.Printf("error: %v\n", err)
	}
	// Write StderrPipe to stdout
	if cmd.headless {
		io.Copy(os.Stdout, rc)
	}

	return nil
}

func (cmd *Command) serialOut() {
	if (cmd.hypervisor == shared.VMwarePlayer || cmd.hypervisor == shared.VMwareWorkstation || cmd.hypervisor == shared.VMwareTest) && cmd.echo {

		t, err := tail.TailFile(fullPath+"/serial.log", tail.Config{Follow: true})
		if err != nil {
			fmt.Println("Err on Tail")
		}
		for line := range t.Lines {
			fmt.Println(line.Text)
		}
	} else if cmd.hypervisor == shared.VirtualBox && cmd.echo {
		t, err := tail.TailFile(vBoxSerial.Name(), tail.Config{Follow: true})
		if err != nil {
			fmt.Println("Err on Tail")
		}
		for line := range t.Lines {
			fmt.Println(line.Text)
		}

	}
}

func (cmd *Command) paramsVMWare(cores, memory, disk, name string, vmplayer bool, numberOfNetworkCards int) ([]string, error) {

	tmpDir := os.TempDir()
	var sub string

	if vmplayer {
		sub = "player"
	} else {
		sub = "workstation"
	}
	name = strings.Replace(name, " ", "", -1)
	fullPath = path.Join(tmpDir, "vorteil", sub, name)
	os.RemoveAll(fullPath)

	vmxName := filepath.Join(fullPath, name+".vmx")

	// just in case
	os.MkdirAll(fullPath, os.ModePerm)

	// delete if we ran it before
	cmD := exec.Command("vmrun", "stop", vmxName)
	cmD.Run()

	cmD = exec.Command("vmrun", "deleteVM", vmxName)
	cmD.Run()

	vmwname = vmxName

	// handles vmx disk name for persist or not persist
	var diskname string
	if cmd.persist != "" {
		diskname = strings.Split(disk, "/")[len(strings.Split(disk, "/"))-1]
	} else {
		diskname = disk
	}

	// generate the vmx string with actual settings
	vmxString := shared.GenerateVMX(cores, memory, diskname, name, fullPath, numberOfNetworkCards)

	err := ioutil.WriteFile(vmxName, []byte(vmxString), os.ModePerm)
	if err != nil {
		return nil, err
	}

	var args []string
	if vmplayer {
		args = []string{vmxName}
	} else {
		args = []string{"start", vmxName}
		if cmd.headless {
			args = append(args, "nogui")
		}
	}

	// Handle Persists
	if cmd.persist != "" {

		old, err := filepath.Abs(cmd.persist)
		if err != nil {
			fmt.Println(err)
		}
		// pers := strings.Split(cmd.persist, "/")
		pers := filepath.SplitList(cmd.persist)
		new := fullPath + "/" + pers[len(pers)-1]

		err = os.Symlink(old, new)
		if err != nil {
			fmt.Println(err)
		}
	}

	return args, nil
}

func (cmd *Command) paramsVirtualBox(cores, memory, disk, name string, numberOfNetworkCards int, mappings string) ([]string, error) {

	// delete vm
	tmpDir := os.TempDir()
	fullPath := path.Join(tmpDir, "vorteil")

	os.RemoveAll(fullPath)

	cmD := exec.Command("VBoxManage", "controlvm", name, "poweroff")
	cmD.Run()

	// if image by specified name already exists, remove image before creating new one
	cmD = exec.Command("VBoxManage", "unregistervm", name)
	cmD.Run()

	// create vm
	cmD = exec.Command("vboxmanage", "createvm", "--basefolder", fullPath, "--name", name, "--register")
	err := cmD.Run()
	if err != nil {
		return nil, err
	}

	args := []string{"modifyvm", name, "--memory", memory, "--acpi", "on", "--ioapic", "on", "--cpus", cores, "--pae", "on"}
	args = append(args, "--longmode", "on", "--largepages", "on", "--chipset", "ich9", "--bioslogofadein", "off")
	args = append(args, "--bioslogofadeout", "off", "--bioslogodisplaytime", "1", "--biosbootmenu", "disabled", "--rtcuseutc", "on")

	if !cmd.headless && cmd.echo {
		vBoxSerial, err = ioutil.TempFile(tmpDir, "vBoxSerial")
		if err != nil {
			fmt.Println(err)
		}
		args = append(args, "--uart1", "0x3F8", "4", "--uartmode1", "file", vBoxSerial.Name())

	}

	if len(mappings) > 0 {
		pmap := strings.Split(mappings, ",")
		for i, p := range pmap {
			pmapIndividual := strings.Split(p, ":")
			x, _ := strconv.Atoi(pmapIndividual[0])

			// TCP ...
			cmD = exec.Command("vboxmanage", "modifyvm", name, "--natpf"+strconv.Itoa(x+1), fmt.Sprintf("nat%dtcp,tcp,,%s,,%s", i, pmapIndividual[1], pmapIndividual[2]))
			err = cmD.Run()

			// UDP ...
			cmD = exec.Command("vboxmanage", "modifyvm", name, "--natpf"+strconv.Itoa(x+1), fmt.Sprintf("nat%dudp,udp,,%s,,%s", i, pmapIndividual[1], pmapIndividual[2]))
			err = cmD.Run()
		}
	}

	for i := 1; i <= numberOfNetworkCards; i++ {
		args = append(args, "--nic"+strconv.Itoa(i), "nat", "--nictype"+strconv.Itoa(i), "82540EM", "--cableconnected"+strconv.Itoa(i), "on")
	}

	cmD = exec.Command("vboxmanage", args...)
	err = cmD.Run()
	if err != nil {
		return nil, err
	}

	cmD = exec.Command("vboxmanage", "storagectl", name, "--name", "SATA", "--add", "sata", "--portcount", "4", "--bootable", "on")
	err = cmD.Run()
	if err != nil {
		return nil, err
	}

	cmD = exec.Command("vboxmanage", "storageattach", name, "--storagectl", "SATA", "--port", "0", "--device", "0", "--type", "hdd", "--medium", disk)
	err = cmD.Run()
	if err != nil {
		return nil, err
	}

	if !cmd.headless {
		args = args[:0]
		args = append(args, "--startvm", name, "--start-running")
		if cmd.debug {
			args = append(args, "--debug")
		}
		args = append(args, "--type", "sdl")
	}

	return args, nil
}

func paramsKVMCloud(cores, memory, disk string, numberOfNetworkCards int, mappings string) []string {
	args := []string{"-cpu", "IvyBridge", "-no-reboot"}
	args = append(args, "-machine", "pc", "-smp", cores, "-m", memory, "-enable-kvm")

	for i := 0; i < numberOfNetworkCards; i++ {
		args = append(args, "-device", "virtio-net-pci,id=network"+strconv.Itoa(i))
	}

	args = append(args, "-device", "virtio-scsi-pci,id=scsi", "-device", "scsi-hd,drive=hd0")
	args = append(args, "-s", "-drive", "if=none,file="+disk+",format=vmdk,id=hd0")

	// split pmap
	if len(mappings) > 0 {
		pmap := strings.Split(mappings, ",")
		for _, p := range pmap {
			pmapIndividual := strings.Split(p, ":")
			args = append(args, "-redir", fmt.Sprintf("tcp:%s::%s", pmapIndividual[0], pmapIndividual[1]))
		}
	}

	return args
}

func (cmd *Command) paramsKVM(cores, memory, disk string, numberOfNetworkCards int, mappings string) []string {
	args := []string{"-cpu", "host", "-no-reboot"}
	args = append(args, "-machine", "q35", "-smp", cores, "-m", memory, "-enable-kvm")

	//Start as headless
	if cmd.headless {
		args = append(args, "-nographic", "-display", "none", "-daemonize")
	}

	if cmd.debug {
		args = append(args, "-s")
	}

	if cmd.hypervisor != shared.KVM {
		// Adds new AHCI SATA drive
		args = append(args, "-drive", "id=disk,file="+disk+",if=none")
		args = append(args, "-device", "ide-drive,drive=disk,bus=ide.0,id=hd0")
	}

	if cmd.echo && !cmd.headless {
		args = append(args, "-serial", "stdio")
	}

	start := 0xa

	if cmd.hypervisor == shared.KVM {
		// KVM

		// virtio scsi hd
		args = append(args, "-device", "virtio-scsi-pci,id=scsi", "-device", "scsi-hd,drive=hd0")
		args = append(args, "-drive", "if=none,file="+disk+",format=vmdk,id=hd0")

		// virtio net pci
		for i := 0; i < numberOfNetworkCards; i++ {

			var card string
			istr := strconv.Itoa(i)
			pmap := strings.Split(mappings, ",")

			if len(mappings) > 0 {

				// loop through port maps and assign where appropriate
				for j := 0; j < len(pmap); j++ {

					pmapInd := strings.Split(pmap[j], ":")
					if pmapInd[0] == istr {
						// TCP ...
						// args = append(args, "-redir", fmt.Sprintf("tcp:%s::%s", pmapInd[1], pmapInd[2]))
						card += fmt.Sprintf(",hostfwd=tcp::%s-:%s", pmapInd[1], pmapInd[2])

						// UDP ...
						// args = append(args, "-redir", fmt.Sprintf("udp:%s::%s", pmapInd[1], pmapInd[2]))
						card += fmt.Sprintf(",hostfwd=udp::%s-:%s", pmapInd[1], pmapInd[2])
					}

				}

			}

			args = append(args, "-netdev", "user,id=network"+istr+card, "-device", "virtio-net-pci,netdev=network"+istr+",id=virtio"+istr+",mac="+fmt.Sprintf("26:10:05:00:00:0%x", start+(i*0x1)))

		}

	} else {
		for i := 0; i < numberOfNetworkCards; i++ {

			istr := strconv.Itoa(i)
			var card string
			pmap := strings.Split(mappings, ",")

			if len(mappings) > 0 {

				// loop through port maps and assign where appropriate
				for j := 0; j < len(pmap); j++ {

					pmapInd := strings.Split(pmap[j], ":")
					if pmapInd[0] == istr {
						// TCP ...
						card += fmt.Sprintf(",hostfwd=tcp::%s-:%s", pmapInd[1], pmapInd[2])

						// UDP ...
						card += fmt.Sprintf(",hostfwd=udp::%s-:%s", pmapInd[1], pmapInd[2])
					}

				}

			}

			args = append(args, "-netdev", "user,id=network"+istr+card, "-device", "e1000,netdev=network"+istr+",mac="+fmt.Sprintf("26:10:05:00:00:0%x", start+(i*0x1)))

		}
	}

	return args
}

func (cmd *Command) addPorts(mappings string) string {
	var out string
	if len(mappings) > 0 {
		pmap := strings.Split(mappings, ",")
		for _, p := range pmap {
			pmapIndividual := strings.Split(p, ":")
			out += ",hostfwd=tcp::" + pmapIndividual[0] + "-:" + pmapIndividual[1]

		}
	}
	return out
}

func (cmd *Command) paramsQemu(cores, memory, disk, debugFile string, numberOfNetworkCards int, mappings string) []string {

	args := []string{"-cpu", "qemu64,+rdtscp,+fsgsbase,+ssse3,+sse4.1,+sse4.2,+x2apic,+invtsc", "-no-reboot"}
	args = append(args, "-machine", "q35", "-smp", cores, "-m", memory)
	args = append(args, "-device", "ahci,id=ahci0", "-device", "ide-drive,bus=ahci0.0,drive=drive-sata0-0-0,id=sata0-0-0")
	args = append(args, "-drive", "if=none,file="+disk+",format=vmdk,id=drive-sata0-0-0")
	if cmd.echo && !cmd.headless {
		args = append(args, "-serial", "stdio")
	}

	start := 0xa

	for i := 0; i < numberOfNetworkCards; i++ {

		istr := strconv.Itoa(i)
		var card string
		pmap := strings.Split(mappings, ",")

		if len(mappings) > 0 {

			// loop through port maps and assign where appropriate
			for j := 0; j < len(pmap); j++ {

				pmapInd := strings.Split(pmap[j], ":")
				if pmapInd[0] == istr {
					card += fmt.Sprintf(",hostfwd=tcp::%s-:%s", pmapInd[1], pmapInd[2])
					card += fmt.Sprintf(",hostfwd=udp::%s-:%s", pmapInd[1], pmapInd[2])
				}

			}

		}

		args = append(args, "-netdev", "user,id=network"+istr+card, "-device", "e1000,netdev=network"+istr+",mac="+fmt.Sprintf("26:10:05:00:00:0%x", start+(i*0x1)))

	}

	if len(debugFile) > 0 {
		args = append(args, "-d", "int,guest_errors,cpu,in_asm,exec", "-D", debugFile)
	}
	if cmd.debug {
		args = append(args, "-s")
	}

	if cmd.headless {
		args = append(args, "-nographic", "-display", "none", "-daemonize")
	}

	return args

}
