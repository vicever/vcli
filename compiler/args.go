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

// CompileArgs contains all fields that can be set to effect the output of a
// compile command.
type CompileArgs struct {
	Binary      string
	Config      string
	Files       string
	Version     string
	Debug       bool
	Destination string
}

func (builder *builder) validateArgs(args *CompileArgs) error {

	// TODO:
	return nil

}
