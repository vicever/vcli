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
	"errors"
	"os"
	"path"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sisatech/sherlock"
)

const (
	envHome = "VORTEIL_HOME"
)

var (
	home = ""

	// Binaries is the internal path to the binary files
	Binaries = "binaries"

	// Icons is the internal path to the app icons
	Icons = "icons"

	// AuxiliaryFiles is the internal path to the auxiliary files
	AuxiliaryFiles = "filetrees"

	// VMLDatabase ...
	VMLDatabase = "vml.db"

	// VMLAppStorage ...
	VMLAppStorage = "apps"

	// Infrastructures ...
	Infrastructures = "infs"

	// Repository ...
	Repository = "repo"

	SafeMode = false
)

// Path returns the absolute path to a file or folder expected to exist within
// the vcli home directory.
func Path(target string) string {

	return home + "/" + target

}

// Initialize checks if the vcli home directory is already setup, and if it
// isn't then it attempts to set it up itself, returning an error if a problem
// occurs.
func Initialize() error {

	// check if an alternative home location has been specified
	if x := os.Getenv(envHome); x != "" {

		home = x

		// if !strings.HasPrefix(home, "/") {
		if !filepath.IsAbs(home) {
			return errors.New(envHome + " environment variable invalid: must not be a relative path")
		}

		if home == "/" {
			return errors.New(envHome + " environment variable invalid: must not be the root directory")
		}

	} else {

		usrHome, err := homedir.Dir()
		if err != nil {
			return err
		}

		if runtime.GOOS == "windows" {
			// windows
			home = path.Join(usrHome, "vorteil")

		} else {
			// linux / mac
			home = path.Join(usrHome, ".vorteil")

		}

		sherlock.Check(os.Setenv(envHome, home))

	}

	// initialize vcli home base directory
	if err := setupDir(home); err != nil {
		return err
	}

	// setup subdirectories
	if err := setupDir(Path(Kernel)); err != nil {
		return err
	}

	if err := setupDir(Path(Binaries)); err != nil {
		return err
	}

	if err := setupDir(Path(Icons)); err != nil {
		return err
	}

	if err := setupDir(Path(AuxiliaryFiles)); err != nil {
		return err
	}

	if err := setupDir(Path(Infrastructures)); err != nil {
		return err
	}

	// create global defaults file
	err := initGlobalDefaults()
	if err != nil {
		return err
	}

	// TODO: setup databases

	return nil

}

// Cleanup safely stores any changes to settings or defaults.
func Cleanup() {

	saveGlobalDefaults()

}
