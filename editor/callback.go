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
package editor

import "strconv"

func (ed *Editor) DisplayCallback(fn func() string, cursor ...int) {

	var key string

	for _, x := range cursor {
		key = key + strconv.Itoa(x) + ";"
	}

	ed.displayFunctions[key] = fn

}

func (ed *Editor) EditCallback(fn func(string) error, tooltip string, fieldVal string, cursor ...int) {

	var key string

	for _, x := range cursor {
		key = key + strconv.Itoa(x) + ";"
	}

	ed.editFunctions[key] = fn
	ed.toolTips[key] = tooltip
	ed.fieldVals[key] = fieldVal
}

//
// func (ed *Editor) FieldToolTips(tooltip string, cursor ...int) {
// 	var key string
//
// 	for _, x := range cursor {
// 		key = key + strconv.Itoa(x) + ";"
// 	}
// 	ed.toolTips[key] = tooltip
// }
