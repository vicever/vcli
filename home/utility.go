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
)

const (
	perm = 0700
)

func setupDir(path string) error {

	info, err := os.Stat(path)

	if err != nil {

		if err == os.ErrInvalid {
			return errors.New(envHome + " environment variable invalid")
		}

		if os.IsPermission(err) {

			return errors.New("vcli home directory cannot be accessed: access denied")

		}

		if os.IsNotExist(err) {

			err = os.MkdirAll(path, perm)
			if err != nil {
				return errors.New("could not create directory: " + err.Error())
			}

			info, err = os.Stat(path)

			if err != nil {
				return errors.New("unexpected error: " + err.Error())
			}

		} else {
			return errors.New("unexpected error: " + err.Error())
		}

	}

	// check is a directory
	if !info.IsDir() {
		return errors.New("target exists and is not a directory: " + path)
	}

	return nil

}
