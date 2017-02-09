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
package converter

import (
	"io/ioutil"
	"os"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/compiler"
)

func ExportSparseVMDK(in Convertible, path, kernel string, debug bool) (*os.File, error) {

	var vmdk *os.File

	err := sherlock.Try(func() {

		var err error

		if path == "" {
			vmdk, err = ioutil.TempFile("", "")
			sherlock.Check(err)
			sherlock.Check(vmdk.Close())
			path = vmdk.Name()
		}

		// create temp dir for files
		tmp, err := ioutil.TempDir("", "")
		sherlock.Check(err)

		sherlock.Check(ExportLoose(in, tmp))

		defer os.RemoveAll(tmp)

		// build sparse vmdk
		out, err := compiler.BuildSparseVMDK(tmp+"/app",
			tmp+"/app.vcfg", tmp+"/fs", kernel, path, debug)
		sherlock.Check(err)

		vmdk, err = os.Open(out)
		sherlock.Check(err)

	})

	return vmdk, err

}
