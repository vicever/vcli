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
package command

import (
	"errors"
	"os"

	"github.com/alecthomas/kingpin"
)

const (
	// help is the string to match if args need to be checked for
	// help flag functionality.
	help = "--help"
)

var (
	// errCommandNode should be thrown whenever a command that only has
	// subcommands is called without a subcommand argument.
	errCommandNode = errors.New("need a valid subcommand (try --help)")
)

// NodeOnlyCheck prints helpful information to the screen if the user has
// selected a command without selecting a valid subcommand.
func NodeOnlyCheck(cmd Command, ctx *kingpin.ParseContext) error {

	for _, x := range os.Args {
		if x == help {
			return nil
		}
	}

	if ctx.SelectedCommand.FullCommand() == cmd.FullCommand() {
		return errCommandNode
	}

	return nil

}
