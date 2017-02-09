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
	"fmt"
	"strings"
)

// Compiler handles all compilation of Vorteil virtual disks
type Compiler struct {
	loud       bool
	kernels    string
	bootLoader string
	trampoline string
}

func New(kernelsDirectory string) (*Compiler, error) {

	compiler := new(Compiler)

	// intialize kernel resources directory
	compiler.kernels = strings.TrimSuffix(kernelsDirectory, "/")
	compiler.bootLoader = compiler.kernels + "/vboot.img"
	compiler.trampoline = compiler.kernels + "/vtramp.img"

	return compiler, nil

}

func (compiler *Compiler) EnableBasicConsoleOutput() {

	compiler.loud = true

}

func (compiler *Compiler) Log(str string, args ...interface{}) {

	if compiler.loud {
		fmt.Printf(str+"\n", args...)
	}

}

func ceiling(arg, interval uint64) uint64 {

	return (arg + interval - 1) / interval

}
