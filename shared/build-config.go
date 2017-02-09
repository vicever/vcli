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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/sisatech/vcli/vml"
)

// BuildConfig holds the individual configuration entries
type BuildConfig struct {
	Name        string    `yaml:"name" json:"name"`
	Description string    `yaml:"description" json:"description"`
	Author      string    `yaml:"author" json:"author"`
	ReleaseDate time.Time `yaml:"release-date" json:"release-date"`
	Version     string    `yaml:"version" json:"version"`
	AppURL      string    `yaml:"url" json:"appurl"`

	App     *BuildAppConfig `yaml:"app,flow" json:"app,flow"`
	Network *NetworkConfig  `yaml:"network,flow" json:"network,flow,omitempty"`
	Disk    *DiskConfig     `yaml:"disk,flow" json:"disk,flow"`

	Redirects *RedirectConfig
	NTP       *NTPConfig
}

// DiskConfig providesd all disk information
type DiskConfig struct {
	FileSystem string `yaml:"file-system" json:"filesystem"`
	MaxFD      int    `yaml:"max-fds" json:"maxfd"`
	DiskSize   int    `yaml:"disk-size" json:"disksize"`
}

// BuildAppConfig contains app specific build information
type BuildAppConfig struct {
	// BinaryType string   `yaml:"type" json:"type"`
	BinaryArgs []string `yaml:"args" json:"binaryargs"`
	SystemEnvs []string `yaml:"envs" json:"systemenvs"`
}

// NetworkCardConfig store config about each network card
type NetworkCardConfig struct {
	IP      string `yaml:"ip" json:"ip,omitempty"`
	Mask    string `yaml:"mask" json:"mask,omitempty"`
	Gateway string `yaml:"gateway" json:"gateway,omitempty"`
}

// NetworkConfig is the general network configuration
type NetworkConfig struct {
	DNS          []string            `yaml:"dns,omitempty" json:"dns,omitempty"`
	NetworkCards []NetworkCardConfig `yaml:"cards,omitempty" json:"cards,omitempty"`
}

// NTPConfig contains hostname and each timeserver
type NTPConfig struct {
	Hostname string   `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	Servers  []string `yaml:"servers,omitempty" json:"servers,omitempty"`
}

// RedirectConfig ...
type RedirectConfig struct {
	Rules []Redirect `yaml:"rules,omitempty" json:"rules,flow,omitempty"`
}

// Redirect
type Redirect struct {
	Src      string `yaml:"src" json:"src,omitempty" nav:"source"`
	Dest     string `yaml:"dest" json:"dest,omitempty" nav:"destination"`
	Protocol string `yaml:"protocol" json:"protocol,omitempty" nav:"protocol"`
}

func VCFGHealthCheck(home, path string) error {

	// Unmarshal the target vcfg file
	// If any of the fields are NULL
	// Replace with anything.
	var repo bool
	buf := make([]byte, 0)
	var err error
	// IF REPO CONFIG FILE, DO APPROPRIATE LOGIC
	if strings.HasPrefix(path, RepoPrefix) {
		repo = true
		path = strings.TrimPrefix(path, RepoPrefix)

		mgr, err := vml.NewTinyRepo(home)
		if err != nil {
			return err
		}

		out, err := mgr.Export(path, "")
		if err != nil {
			return err
		}

		buf, err = ioutil.ReadAll(out.Config())
		if err != nil {
			return err
		}
	}

	path = strings.TrimSuffix(path, ".vcfg")
	path += ".vcfg"

	// Prompt user to fix
	// reader := bufio.NewReader(os.Stdin)
	var f *BuildConfig
	f = new(BuildConfig)

	if !repo {
		buf, err = ioutil.ReadFile(path)
		if err != nil {
			return err
		}
	}

	err = json.Unmarshal(buf, f)
	if err != nil {
		return err
	}
	nullsFound := false

	// Any null fields detected ...?
	if f.App == nil || f.Network == nil || f.Disk == nil || f.NTP == nil || f.Redirects == nil {
		nullsFound = true
	}

	if nullsFound {
		if !repo {
			fmt.Println(Catenate(`The config file '` + path + `' is from an older
				version of VCLI and must be fixed before the current operation
				can continue. Should VCLI automatically fix the config file? [y/n]`))
			var input string
			for {
				n, err := fmt.Scanln(&input)
				if err != nil || n != 1 {
					continue
				}
				if input == "n" {
					return errors.New("user opted out of vcfg file repair. Aborting operation.")
				} else {
					if input == "y" {
						break
					}
				}
			}

			// CHECK IF EACH ELEMENT WITHIN THE VCFG IS NOT NIL
			// Redirects ...
			if f.Redirects == nil {
				f.Redirects = new(RedirectConfig)
				f.Redirects.Rules = make([]Redirect, 0)
			}
			// NTP ...
			if f.NTP == nil {
				f.NTP = new(NTPConfig)
				f.NTP.Hostname = ""
				f.NTP.Servers = make([]string, 0)
			}
			// App ...
			if f.App == nil {
				f.App = new(BuildAppConfig)
				f.App.BinaryArgs = make([]string, 0)
				f.App.SystemEnvs = make([]string, 0)
			}
			// Network cards ...
			if f.Network == nil {
				f.Network = new(NetworkConfig)
				f.Network.NetworkCards = make([]NetworkCardConfig, 0)
			}
			// Disk ...
			if f.Disk == nil {
				f.Disk = new(DiskConfig)
			}

			out, err := json.MarshalIndent(f, "", "  ")
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(path, out, 0666)
			if err != nil {
				return err
			}
		} else {
			return errors.New("The .vcfg in this repository is out of date.")
		}
	}

	return nil

}
