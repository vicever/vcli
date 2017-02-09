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

import "github.com/gizak/termui"

type gui struct {
	header  *termui.Par
	crumbs  *termui.Par
	nav     *termui.List
	content *termui.Par
	footer  *termui.Par
	input   *termui.Par
	IBody   *termui.Grid
}

func (ed *Editor) initGUI() {

	ed.Log("INIT GUI")

	// header
	ed.content.title = "TITLE"
	ed.gui.header = termui.NewPar("")
	ed.gui.header.Border = false
	ed.gui.header.TextBgColor = termui.ColorWhite
	ed.gui.header.TextFgColor = termui.ColorBlack
	ed.gui.header.Bg = termui.ColorWhite
	ed.gui.header.Height = 1

	// crumbs
	ed.gui.crumbs = termui.NewPar("")
	ed.gui.crumbs.PaddingLeft = 1
	ed.gui.crumbs.Border = false
	ed.gui.crumbs.Height = 1

	// navigation
	ed.gui.nav = termui.NewList()
	ed.gui.nav.Height = 21

	// content
	ed.gui.content = termui.NewPar("CONTENT")
	ed.gui.content.Height = 21

	// footer
	ed.content.footer = "↑↓: select   ENTER: enter   ESC: back   s: save   q: quit"
	ed.gui.footer = termui.NewPar("")
	ed.gui.footer.TextFgColor = termui.ColorBlue
	ed.gui.footer.Border = false
	ed.gui.footer.TextBgColor = termui.ColorWhite
	ed.gui.footer.Bg = termui.ColorWhite
	ed.gui.footer.Height = 1

	termui.Body.AddRows(
		termui.NewRow(termui.NewCol(12, 0, ed.gui.header)),
		termui.NewRow(termui.NewCol(12, 0, ed.gui.crumbs)),
		termui.NewRow(termui.NewCol(3, 0, ed.gui.nav), termui.NewCol(9, 0, ed.gui.content)),
		termui.NewRow(termui.NewCol(12, 0, ed.gui.footer)),
	)

	ed.gui.input = termui.NewPar("")
	ed.gui.input.Height = termui.TermHeight() / 2
	ed.gui.input.Width = termui.TermWidth() / 2
	ed.gui.input.SetX(ed.gui.input.Width / 2)
	ed.gui.input.SetY(ed.gui.input.Height / 2)
	ed.gui.input.BorderLabel = "Test"
	ed.gui.input.BorderLabelFg = termui.ColorWhite
	ed.gui.input.PaddingTop = 1
	ed.gui.input.PaddingBottom = 1
	ed.gui.input.PaddingLeft = 1
	ed.gui.input.PaddingRight = 1

	ed.gui.IBody = termui.NewGrid()
	ed.gui.IBody.AddRows(
		termui.NewRow(termui.NewCol(12, 0, ed.gui.header)),
		termui.NewRow(termui.NewCol(12, 0, ed.gui.crumbs)),
		termui.NewRow(termui.NewCol(12, 0, ed.gui.input)),
		termui.NewRow(termui.NewCol(12, 0, ed.gui.footer)),
	)
}

func (ed *Editor) padTitle(width int) {

	l := len(ed.content.title)
	remainder := width - l
	left := remainder / 2

	x := ""
	for i := 0; i < left; i++ {
		x = x + " "
	}
	x = x + ed.content.title

	ed.gui.header.Text = x

}

func (ed *Editor) padFooter(width int) {

	l := len(ed.content.footer)
	remainder := width - l
	left := remainder / 2

	x := ""
	for i := 0; i < left; i++ {
		x = x + " "
	}
	x = x + ed.content.footer

	ed.gui.footer.Text = x

}

func (ed *Editor) refresh() {

	ed.render(termui.Event{})

}
