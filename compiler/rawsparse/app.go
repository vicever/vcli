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
package rawsparse

import (
	"io/ioutil"
	"os"
)

func (build *builder) writeApp() error {

	path := build.args.Binary

	app, err := os.Open(path)
	if err != nil {
		return err
	}
	defer app.Close()

	in, err := ioutil.ReadAll(app)
	if err != nil {
		return err
	}

	_, err = build.content.file.WriteAt(in, int64(build.content.app.first*SectorSize))
	if err != nil {
		return err
	}

	build.log("Writing app at LBAs: %d - %d", build.content.app.first, build.content.app.last)

	return nil

}
