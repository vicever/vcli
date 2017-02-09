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

import (
	"archive/zip"
	"debug/elf"

	"github.com/sisatech/sherlock"
)

func IsZip(path string) bool {

	err := sherlock.Try(func() {

		r, err := zip.OpenReader(path)
		sherlock.Check(err)
		r.Close()

	})

	return err == nil

}

func IsELF(path string) bool {

	err := sherlock.Try(func() {

		f, err := elf.Open(path)
		sherlock.Check(err)
		f.Close()

	})

	return err == nil

}
