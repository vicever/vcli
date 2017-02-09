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
package home

import (
	"io/ioutil"
	"strings"
)

const (
	// Kernel is the internal path to kernel files
	Kernel = "kernel"
)

// ValidLocalKernel returns true if the kernel exists locally.
func ValidLocalKernel(kernel string) bool {

	vals := ListLocalKernels()
	for _, x := range vals {

		if x == kernel {
			return true
		}

	}

	return false

}

// ListLocalKernels returns an array of kernel versions found locally.
func ListLocalKernels() []string {

	var ret []string

	ls, err := ioutil.ReadDir(Path(Kernel))
	if err != nil {
		return ret
	}

	for _, x := range ls {

		if strings.HasPrefix(x.Name(), "vkernel-PROD-") &&
			!strings.HasSuffix(x.Name(), ".asc") {

			version := strings.TrimPrefix(x.Name(), "vkernel-PROD-")
			version = strings.TrimSuffix(version, ".img")
			ret = append(ret, version)

		}

	}

	return ret

}
