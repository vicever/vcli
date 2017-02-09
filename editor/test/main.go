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
package main

import (
	"fmt"
	"os"

	"github.com/sisatech/vcli/editor"
)

type MyData struct {
	ALPHA   string `woot:"allooo"`
	BRAVO   int
	CHARLIE struct {
		DELTA int32
		ECHO  string
	}
}

var (
	mydata = MyData{
		ALPHA: "alpha",
		BRAVO: 42,
		CHARLIE: struct {
			DELTA int32
			ECHO  string
		}{
			DELTA: 16,
			ECHO:  "uwot",
		},
	}
)

func main() {

	ed, err := editor.New(mydata)

	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer ed.Cleanup()

	ed.Title("MY FANCY TITLE")
	ed.DisplayCallback(alphaDisplay, 0)
	ed.EditCallback(alphaEdit, 0)

	ed.Run()

	ed.Log("DONE RUNNING")

}

func alphaDisplay() string {
	return mydata.ALPHA
}

func alphaEdit(input string) error {

	mydata.ALPHA = input
	return nil

}
