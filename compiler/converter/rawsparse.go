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
	"path"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/compiler"
)

func ExportRAWSparse(in Convertible, pathStr, kernel string, debug bool) error {
	var img *os.File
	err := sherlock.Try(func() {

		var err error

		if pathStr == "" {
			img, err = ioutil.TempFile("", "")
			sherlock.Check(err)
			sherlock.Check(img.Close())
			pathStr = img.Name()
		}

		// create temp dir for files
		tmp, err := ioutil.TempDir("", "")
		sherlock.Check(err)

		sherlock.Check(ExportLoose(in, tmp))

		defer os.RemoveAll(tmp)

		// build raw sparse
		out, err := compiler.BuildRawSparse(tmp+"/app",
			tmp+"/app.vcfg", tmp+"/fs", kernel, pathStr, debug)
		sherlock.Check(err)

		_, out = path.Split(out)
		img, err = os.Open(out)
		sherlock.Check(err)

	})
	// defer os.Remove(img.Name())
	return err

}
