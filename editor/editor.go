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

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gizak/termui"
)

type Editor struct {
	log              *os.File
	obj              interface{}
	cursor           []int
	crumbs           []string
	nav              []string
	gui              gui
	content          content
	cleaned          bool
	displayFunctions map[string]func() string
	editFunctions    map[string]func(string) error
	toolTips         map[string]string
	fieldVals        map[string]string
	gettingInput     bool
	saveOrQuit       error
	inputCurs        int
}

func New(o interface{}) (*Editor, error) {

	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}

	err = termui.Init()
	if err != nil {
		return nil, err
	}

	ed := new(Editor)
	ed.log = f
	ed.obj = o
	ed.displayFunctions = make(map[string]func() string)
	ed.editFunctions = make(map[string]func(string) error)
	ed.toolTips = make(map[string]string)
	ed.fieldVals = make(map[string]string)

	ed.initRenderSettings()
	ed.initKeybinds()
	ed.initGUI()
	ed.initCursor()

	return ed, nil

}

func (ed *Editor) Run() error {

	ed.Log("RUN")
	ed.render(termui.Event{})
	ed.displayContent()
	termui.Loop()
	ed.Log("ERROR SAVEORQUIT IS: %v", ed.saveOrQuit)
	return ed.saveOrQuit

}

func (ed *Editor) Cleanup() {

	if ed.log != nil {
		ed.log.Close()
	}

	if !ed.cleaned {
		ed.cleaned = true
		termui.Close()
	}

}

func (ed *Editor) Log(msg string, args ...interface{}) {

	fmt.Fprintf(ed.log, msg+"\n", args...)

}
