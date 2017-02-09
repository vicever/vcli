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
	"os"
	"strconv"
	"strings"

	"github.com/gizak/termui"
)

func (ed *Editor) initRenderSettings() {
	// ATTENTION: Have moved this handler to initKeybinds so that it is not
	// broken by the ResetHandlers call made when getting user input!!
	// window automatic resizing
	termui.Handle("/sys/wnd/resize", ed.render)

}

func (ed *Editor) render(e termui.Event) {

	if ed.gettingInput {
		// getInput
		termui.Body.Width = termui.TermWidth()
		ed.padTitle(termui.Body.Width)
		ed.padFooter(termui.Body.Width)
		ed.gui.input.Height = termui.TermHeight() / 2
		ed.gui.input.Width = termui.TermWidth() / 2
		ed.gui.input.SetX(ed.gui.input.Width / 2)
		ed.gui.input.SetY(ed.gui.input.Height / 2)
		termui.Clear()
		termui.Render(ed.gui.IBody)
	} else {

		// ed.initKeybinds()

		h := termui.TermHeight()
		ed.gui.nav.Height = h - 3
		ed.gui.content.Height = h - 3

		termui.Body.Width = termui.TermWidth()
		ed.padTitle(termui.Body.Width)
		ed.padFooter(termui.Body.Width)
		termui.Body.Align()
		termui.Clear()

		ed.displayContent()

		termui.Render(termui.Body)
	}

}

func (ed *Editor) getInput() (string, error) {

	termui.DefaultEvtStream.ResetHandlers()
	termui.Handle("/sys/kbd/<left>", func(termui.Event) {
		if ed.inputCurs > 0 {
			ed.inputCurs--
		}
	})

	ed.gettingInput = true
	ed.content.footer = "ENTER: enter   ESC: back"
	ed.padFooter(termui.Body.Width)

	termui.Clear()
	termui.Render(ed.gui.IBody)

	i := 0

	var curs string
	for _, x := range ed.cursor {
		curs = curs + strconv.Itoa(x) + ";"
	}

	// str is equal to the value of the selected field
	str := ed.fieldVals[curs]

	// get user input 1 byte at a time
	var logbyte = make([]byte, 3, 3)
	for {
		cursOS := len(str) - ed.inputCurs
		ed.inputCurs = len(str) - cursOS
		curstr := ed.addCursor(str)

		termui.Clear()
		termui.Render(ed.gui.IBody)

		tooltip, _ := ed.toolTips[curs]
		ed.gui.input.Text = tooltip + "\n\n" + curstr

		termui.Render(ed.gui.IBody)
		b := make([]byte, 3, 3)
		keyLen, _ := os.Stdin.Read(b)
		ed.Log("%v", keyLen)
		i++
		s := string(b[0])
		nB := byte(13)
		delByte := byte(127)
		escByte := byte(27)
		arrKeys := byte(79)
		leftArr := byte(68)
		rightArr := byte(67)

		logbyte = b

		if b[0] == nB {
			ed.fieldVals[curs] = strings.TrimSpace(str)
			termui.Render(ed.gui.IBody)
			ed.Log("%v", logbyte)
			break
		} else if b[0] == delByte {
			// backspace
			if len(str) > 0 {
				str = ed.delInput(str)
			}
			ed.Log("%v", logbyte)
		} else if b[0] == escByte {
			if b[1] == arrKeys {
				if b[2] == leftArr {
					if ed.inputCurs > 0 {
						ed.Log("%v", logbyte)
						ed.inputCurs--
						curstr = ed.addCursor(str)
					}
				} else if b[2] == rightArr {
					if ed.inputCurs < len(curstr)-1 {
						ed.Log("%v", logbyte)
						ed.inputCurs++
						curstr = ed.addCursor(str)
					}
				}
			} else {
				// esc
				ed.Log("%v", logbyte)
				str = ""
				termui.Render(ed.gui.IBody)
				ed.gettingInput = false
				ed.content.footer = "↑↓: select   ENTER: enter   ESC: back   s: save   q: quit"
				ed.padFooter(termui.Body.Width)
				ed.refresh()
				// termui.Render(ed.gui.IBody)
				return str, errors.New("User manually terminated the process.")
			}
		} else {
			ed.Log("%v", logbyte)
			str = ed.addInput(str, s)
		}
		termui.Render(ed.gui.IBody)
	}

	ed.Log("%v", logbyte)

	ed.gettingInput = false
	ed.content.footer = "↑↓: select   ENTER: enter   ESC: back   s: save   q: quit"

	ed.padFooter(termui.Body.Width)
	ed.refresh()
	return str, nil
}

func (ed *Editor) delInput(str string) string {
	var foo string
	letters := strings.Split(str, "")
	for i := 0; i < len(letters); i++ {
		if ed.inputCurs > 0 {
			if i != ed.inputCurs-1 {
				foo += letters[i]
			}
		} else {
			return str
		}
	}
	ed.inputCurs--
	return foo
}

func (ed *Editor) addInput(str string, s string) string {
	var foo string
	letters := strings.Split(str, "")
	for i := 0; i < len(letters)+1; i++ {
		if i == ed.inputCurs {
			foo += s
		}
		if i < len(letters) {
			foo += letters[i]
		}
	}

	ed.inputCurs++
	return foo
}

func (ed *Editor) addCursor(str string) string {
	var foo string
	letters := strings.Split(str, "")
	for i := 0; i < len(letters)+1; i++ {
		if i == ed.inputCurs {
			foo += "|"
		}
		if i < len(letters) {
			foo += letters[i]
		}
	}
	return foo

}

func (ed *Editor) cleanup() {

	termui.Close()

}
