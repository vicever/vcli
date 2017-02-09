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
package compiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"

	"github.com/cavaliercoder/grab"
	"github.com/sisatech/vcli/automation"
	"github.com/sisatech/vcli/home"
)

// DownloadFile ...
type DownloadFile struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Repo    string    `json:"repo"`
	Package string    `json:"package"`
	Version string    `json:"version"`
	Owner   string    `json:"owner"`
	Created time.Time `json:"created"`
	Size    int       `json:"size"`
	SHA1    string    `json:"sha1"`
	SHA256  string    `json:"sha256"`
}

// ListTuple ...
type ListTuple struct {
	Name    string
	Created time.Time
	Local   bool
}

const (
	kernelAPIURL      = "https://api.bintray.com/packages/sisatech/vorteil/%s/files"
	kernelDownloadURL = "https://dl.bintray.com/sisatech/vorteil/"
)

// ListKernels lists all kernels
func ListKernels() ([]*ListTuple, error) {

	home.GlobalDefaults.KnownKernelVersions = []string{}

	var ret []*ListTuple

	lookup, err := LookupListKernels()
	if err != nil {
		return nil, err
	}

	for _, x := range lookup {

		if strings.HasSuffix(x.Name, ".asc") {
			continue
		}

		if strings.Contains(x.Name, "DEBUG") {
			continue
		}

		version := strings.TrimPrefix(x.Name, "vkernel-PROD-")
		version = strings.TrimSuffix(version, ".img")

		// TODO: robust validation of local files
		var isLocal bool
		local := home.Path(home.Kernel + "/" + x.Name)
		_, err := os.Stat(local)
		if err == nil {
			isLocal = true
		}

		home.GlobalDefaults.KnownKernelVersions = append(home.GlobalDefaults.KnownKernelVersions, version)

		ret = append(ret, &ListTuple{
			Name:    version,
			Created: x.Created,
			Local:   isLocal,
		})

	}

	return ret, nil

}

// LookupListKernels lists all kernels available on bintray.
func LookupListKernels() ([]DownloadFile, error) {

	// TODO: cleanup and/or remove this function

	response, err := http.Get(fmt.Sprintf(kernelAPIURL, "vkernel"))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	jsonList := new(bytes.Buffer)
	_, err = io.Copy(jsonList, response.Body)
	if err != nil {
		return nil, err
	}

	var dat []DownloadFile

	if err := json.Unmarshal(jsonList.Bytes(), &dat); err != nil {
		return nil, err
	}

	return dat, nil

}

// LookupListVorteilFiles lists all kernels available on bintray.
func LookupListVorteilFiles(dlType string) ([]DownloadFile, error) {

	// TODO: cleanup and/or remove this function

	response, err := http.Get(fmt.Sprintf(kernelAPIURL, dlType))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	jsonList := new(bytes.Buffer)
	_, err = io.Copy(jsonList, response.Body)
	if err != nil {
		return nil, err
	}

	var dat []DownloadFile

	if err := json.Unmarshal(jsonList.Bytes(), &dat); err != nil {
		return nil, err
	}

	return dat, nil

}

// DownloadVorteilFile ...
func DownloadVorteilFile(dlFile, dlType string) error {

	// TODO: cleanup and/or remove this function

	kernels, err := LookupListVorteilFiles(dlType)

	if err != nil {
		return err
	}

	for _, k := range kernels {

		// found kernel: download and check signature
		if k.Name == dlFile {

			urls := make([]string, 2)
			urls[0] = kernelDownloadURL + k.Path
			urls[1] = kernelDownloadURL + k.Path + ".asc"

			// delete exisiting files, if downloaded already
			os.Remove(os.TempDir() + "/" + dlFile)
			os.Remove(os.TempDir() + "/" + dlFile + ".asc")
			respch, err := grab.GetBatch(2, os.TempDir(), urls...)
			if err != nil {
				return err
			}
			t := time.NewTicker(10 * time.Millisecond)
			completed := 0
			completedSuccessful := 0
			inProgress := 0
			var responses []*grab.Response

			for completed < len(urls) {
				select {
				case resp := <-respch:
					if resp != nil {
						responses = append(responses, resp)
					}

				case <-t.C:
					if inProgress > 0 {
						fmt.Printf("\033[%dA\033[K", inProgress)
					}

					for i, resp := range responses {
						if resp != nil && resp.IsComplete() {
							if resp.Error != nil {
								fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", resp.Request.URL(), resp.Error)
							} else {
								completedSuccessful++
								fmt.Printf("Finished %s %d / %d bytes (%d%%)\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
							}

							responses[i] = nil
							completed++
						}
					}

					inProgress = 0
					for _, resp := range responses {
						if resp != nil {
							inProgress++
							fmt.Printf("Downloading %s %d / %d bytes (%d%%)\033[K\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
						}
					}
				}
			}

			t.Stop()

			if completedSuccessful < 2 {
				return fmt.Errorf("Could not download file %s", dlFile)
			}

			vorteilHome := strings.TrimSuffix(os.Getenv("VORTEIL_HOME"), "/")

			krrData, err := build.Asset("vorteil.gpg")
			if err != nil {
				return err
			}

			krr := bytes.NewReader(krrData)

			sig, err := os.Open(os.TempDir() + "/" + dlFile + ".asc")
			defer sig.Close()
			if err != nil {
				return err
			}
			target, err := os.Open(os.TempDir() + "/" + dlFile)
			defer target.Close()
			if err != nil {
				return err
			}
			kr, err := openpgp.ReadArmoredKeyRing(krr)
			if err != nil {
				return err
			}

			fmt.Print("Verifying signature")
			_, err = openpgp.CheckArmoredDetachedSignature(kr, target, sig)
			if err != nil {
				return err
			}
			fmt.Print(": Ok\n")

			// copy to kernel folder. Reading full file. Its max 2MB
			data, err := ioutil.ReadFile(os.TempDir() + "/" + dlFile)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(vorteilHome+"/kernel/"+dlFile, data, 0644)
			if err != nil {
				return err
			}

		}

	}

	return nil
}
