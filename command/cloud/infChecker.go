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
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/sisatech/vcli/home"
)

type infType struct {
	Type string
}

func checkInfType(file string) (string, error) {

	// Check infrastructure exists
	fullPath := home.Path(home.Infrastructures) + "/" + file
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", errors.New("infrastructure" + file + "not found")
	}

	// Attempt to unmarshal given Inf file to check inf type ...
	inf := new(infType)

	buf, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	err = yaml.Unmarshal(buf, inf)
	if err != nil {
		return "", err
	}

	// Return detected type
	return inf.Type, nil
}
