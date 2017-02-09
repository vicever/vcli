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

type content struct {
	title  string
	footer string
}

func (ed *Editor) Title(title string) {

	ed.content.title = title
	ed.gui.header.Text = title

}

func (ed *Editor) displayContent() {

	var key string

	for _, x := range ed.cursor {
		key = key + strconv.Itoa(x) + ";"
	}

	fn, ok := ed.displayFunctions[key]
	if !ok {
		ed.gui.content.Text = ""
	} else {
		ed.gui.content.Text = fn()
	}

}
