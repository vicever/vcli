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
package gcpInf

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/editor"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
)

type Infrastructure struct {
	// ClientID string
	Bucket string
	Zone   string
}

var (
	x    []byte
	orig = shared.GCPInfrastructure{
		Bucket: "",
		Zone:   "",
		Type:   "",
		Key:    x,
	}

	infra = Infrastructure{
		Bucket: "",
		Zone:   "",
	}
)

func New(name, key string, edit bool) error {

	if edit {
		// unmarshal file and apply to 'infra'
		buf, err := ioutil.ReadFile(name)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(buf, &orig)
		if err != nil {
			return err
		}

		infra.Bucket = orig.Bucket
		infra.Zone = orig.Zone

	}

	ed, err := editor.New(infra)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer ed.Cleanup()
	ed.Title("VCLI - Infrastructure Client (GOOGLE)")

	// Top Level Callbacks
	ed.DisplayCallback(dBucket, 0)
	ed.DisplayCallback(dZone, 1)
	ed.EditCallback(eBucket, "Name of the Google Cloud Storage bucket to access.", infra.Bucket, 0)
	ed.EditCallback(eZone, "Name of the zone where the bucket is located. (eg. global)", infra.Zone, 1)

	err = ed.Run()
	if err == nil {
		var err2 error
		name, err2 = save(name, key, edit)
		if err2 != nil {
			return err2
		}

	}
	ed.Log("DONE RUNNING")

	return err
}

func save(name, key string, edit bool) (string, error) {

	err := sherlock.Try(func() {

		finalStruct := new(shared.GCPInfrastructure)
		finalStruct.Bucket = strings.ToLower(infra.Bucket)
		finalStruct.Zone = strings.ToLower(infra.Zone)
		finalStruct.Type = shared.GCPInf

		if edit {
			finalStruct.Key = orig.Key
		} else {
			// Read in contents of key file as []byte
			// Save value to finalStruct.Key
			buf, err := ioutil.ReadFile(key)
			sherlock.Check(err)

			finalStruct.Key = buf

		}

		out, err := yaml.Marshal(finalStruct)
		sherlock.Check(err)

		if name == "" {
			err = ioutil.WriteFile(home.Path(home.Infrastructures)+"/default", out, 0666)
			sherlock.Check(err)
			name = "default"
		} else {
			err = ioutil.WriteFile(name, out, 0666)
			sherlock.Check(err)
		}
	})

	return name, err
}

func dBucket() string {
	return infra.Bucket
}

func eBucket(str string) error {
	infra.Bucket = str
	return nil
}

func dZone() string {
	return infra.Zone
}

func eZone(str string) error {
	infra.Zone = str
	return nil
}
