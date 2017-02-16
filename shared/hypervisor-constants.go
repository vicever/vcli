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
package shared

import "os/exec"

// Hypervisor constants
const (
	BinaryQEMU              = "qemu-system-x86_64"
	BinaryVirtualBox        = "VirtualBox"
	BinaryVMwareWorkstation = "vmrun"
	BinaryVMwarePlayer      = "vmplayer"

	QEMU              = "QEMU"
	KVM               = "KVM"
	VirtualBox        = "VIRTUALBOX"
	VMwareWorkstation = "VMWARE"
	VMwarePlayer      = "VMWARE_PLAYER"

	KVMClassic = "KVM_CLASSIC"
	VMwareTest = "VMWARE_TEST"
)

// ListDetectedHypervisors returns an array of hypervisor ID strings matching
// vcli supported hypervisors that have been detected on the system.
func ListDetectedHypervisorsWithHidden() []string {

	var ret []string

	target, _ := exec.LookPath(BinaryQEMU)
	if target != "" {
		ret = append(ret, KVM)
		ret = append(ret, QEMU)
		ret = append(ret, KVMClassic)
	}

	target, _ = exec.LookPath(BinaryVirtualBox)
	if target != "" {
		ret = append(ret, VirtualBox)
	}

	target, _ = exec.LookPath(BinaryVMwarePlayer)
	if target != "" {
		ret = append(ret, VMwarePlayer)
	}

	target, _ = exec.LookPath(BinaryVMwareWorkstation)
	if target != "" {
		ret = append(ret, VMwareWorkstation)
		ret = append(ret, VMwareTest)
	}

	return ret

}

// ListDetectedHypervisors returns an array of hypervisor ID strings matching
// vcli supported hypervisors that have been detected on the system.
func ListDetectedHypervisors() []string {

	var ret []string

	target, _ := exec.LookPath(BinaryQEMU)
	if target != "" {
		ret = append(ret, KVM)
		ret = append(ret, QEMU)
	}

	target, _ = exec.LookPath(BinaryVirtualBox)
	if target != "" {
		ret = append(ret, VirtualBox)
	}

	target, _ = exec.LookPath(BinaryVMwarePlayer)
	if target != "" {
		ret = append(ret, VMwarePlayer)
	}

	target, _ = exec.LookPath(BinaryVMwareWorkstation)
	if target != "" {
		ret = append(ret, VMwareWorkstation)
	}

	return ret

}

// ValidHypervisor returns true if the provided string is a valid choice.
func ValidHypervisor(hypervisor string) bool {

	vals := ListDetectedHypervisors()
	for _, x := range vals {

		if x == hypervisor {
			return true
		}

	}

	return false

}

// ValidHypervisorWithHidden
func ValidHypervisorWithHidden(hypervisor string) bool {

	vals := ListDetectedHypervisorsWithHidden()
	for _, x := range vals {

		if x == hypervisor {
			return true
		}

	}

	return false

}
