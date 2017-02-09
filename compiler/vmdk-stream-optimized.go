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
	"os"

	"github.com/sisatech/vcli/compiler/vmdk"
	"github.com/sisatech/vcli/home"
)

// BuildStreamOptimizedVMDK returns the name of a compiled stream-optimized
// .vmdk disk image within a temporary folder. The caller should move the file
// to a non-temporary location.
func BuildStreamOptimizedVMDK(binary, config, files, kernel, destination string, debug bool) (*os.File, error) {

	err := FullValidation(binary, config, files, "", kernel)
	if err != nil {
		return nil, err
	}

	// TODO: implement
	env, err := vmdk.NewEnvironment(home.Path(home.Kernel))
	if err != nil {
		return nil, err
	}

	env.GrantWriteAccess()
	env.LogToStdout()

	f, err := env.BuildStreamOptimized(&vmdk.BuildArgs{
		Binary:      binary,
		Config:      config,
		Files:       files,
		Kernel:      kernel,
		Debug:       debug,
		Destination: destination,
	})
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return f, err
	}

	return f, nil

}
