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
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/gizak/termui"
)

func (ed *Editor) initKeybinds() {

	// automatic window resizing
	termui.Handle("/sys/wnd/resize", ed.render)

	// quit
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		ed.saveOrQuit = errors.New("User quit without saving changes.")
		termui.StopLoop()
	})

	// save and quit
	termui.Handle("/sys/kbd/s", func(termui.Event) {
		ed.saveOrQuit = nil
		termui.StopLoop()
	})

	// 0-9
	termui.Handle("/sys/kbd/0", ed.numKey)
	termui.Handle("/sys/kbd/1", ed.numKey)
	termui.Handle("/sys/kbd/2", ed.numKey)
	termui.Handle("/sys/kbd/3", ed.numKey)
	termui.Handle("/sys/kbd/4", ed.numKey)
	termui.Handle("/sys/kbd/5", ed.numKey)
	termui.Handle("/sys/kbd/6", ed.numKey)
	termui.Handle("/sys/kbd/7", ed.numKey)
	termui.Handle("/sys/kbd/8", ed.numKey)
	termui.Handle("/sys/kbd/9", ed.numKey)

	// enter
	termui.Handle("/sys/kbd/<enter>", ed.enterKey)

	// escape
	termui.Handle("/sys/kbd/<escape>", ed.escapeKey)

	// arrows
	termui.Handle("/sys/kbd/<down>", ed.downArrow)
	termui.Handle("/sys/kbd/<up>", ed.upArrow)

	termui.Handle("/sys/kbd", ed.logKey)

}

func (ed *Editor) downArrow(e termui.Event) {
	x := ed.cursor[len(ed.cursor)-1]
	if x >= len(ed.nav)-1 {
		return
	}
	ed.cursor[len(ed.cursor)-1] = x + 1
	ed.displayContent()
	ed.updateCursorGraphic()
	ed.refresh()
}

func (ed *Editor) upArrow(e termui.Event) {
	x := ed.cursor[len(ed.cursor)-1]
	if x == 0 {
		return
	}
	ed.cursor[len(ed.cursor)-1] = x - 1
	ed.displayContent()
	ed.updateCursorGraphic()
	ed.refresh()
}

func (ed *Editor) escapeKey(e termui.Event) {

	// Go back 1 level in the GUI
	ed.Log("%v", e)
	if len(ed.cursor) > 1 {
		k := len(ed.cursor) - 1
		ed.cursor = ed.cursor[:k]
	}

	// Update Breadcrumbs
	foo := strings.Split(ed.gui.crumbs.Text, " > ")
	if len(foo) > 1 {
		x := len(foo) - 1
		foo = foo[:x]
		bar := ""

		for i := 0; i < len(foo); i++ {
			if i != 0 {
				bar += " > " + foo[i]
			} else {
				bar += foo[i]
			}
		}
		ed.gui.crumbs.Text = bar
	}

	ed.displayContent()
	ed.fields()
	ed.render(termui.Event{})
}

func (ed *Editor) enterKey(e termui.Event) {

	ed.Log("%v", e)

	var curs string
	for _, x := range ed.cursor {
		curs = curs + strconv.Itoa(x) + ";"
	}

	callback, ok := ed.editFunctions[curs]
	if !ok {
		typ := ed.getType()
		par := ed.getParent()
		ed.Log("%v ||| %v", ed.cursor, typ.String())
		if typ.Kind() != reflect.Struct && typ.Kind() != reflect.Slice {
			// User wants to edit this field
			ed.gui.content.Text = "This is not struct or slice."
		} else {
			// Descend into this struct/array
			ed.cursor = append(ed.cursor, 0)

			ed.Log("ENTER KEY PRESSED. %v", ed.cursor)

			y := len(ed.cursor) - 2
			name := par.Field(ed.cursor[y]).Name
			f, boo := par.FieldByName(name)
			var s string
			if !boo {
				s = name
			} else {
				s = f.Tag.Get("nav")
			}
			// ed.gui.crumbs.Text += " > " + par.Field(ed.cursor[y]).Name
			ed.gui.crumbs.Text += " > " + s
			ed.fields()
			ed.render(termui.Event{})

		}
	} else {

		// reset input cursor
		if len(ed.fieldVals[curs]) > 0 {
			ed.inputCurs = len(ed.fieldVals[curs])
		} else {
			ed.inputCurs = 0
		}

		tag := ed.getTag()
		t, ok := tag.Tag.Lookup("nav")
		if !ok {
			t = ed.getName()
		}
		ed.gui.input.BorderLabel = " Enter new value for: " + t + " "
		str, err := ed.getInput()
		if err == nil {
			callback(str)
		}
	}

	ed.initKeybinds()
	ed.displayContent()
	ed.refresh()

}

func (ed *Editor) logKey(e termui.Event) {

	ed.Log("%v", e)
	ed.refresh()

}

func (ed *Editor) numKey(e termui.Event) {

	num, err := strconv.Atoi(e.Data.(termui.EvtKbd).KeyStr)
	if err != nil {
		panic(err)
	}

	ed.Log("KEYPRESS: %v", num)

	if num == 0 {
		num = 10
	}

	num--

	if num >= len(ed.nav) {
		return
	}

	ed.cursor[len(ed.cursor)-1] = num

	ed.displayContent()

	ed.updateCursorGraphic()
	ed.refresh()

}
