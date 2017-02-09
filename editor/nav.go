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

import "reflect"

func (ed *Editor) initCrumbs() {

	typ := reflect.TypeOf(ed.obj)
	ed.crumbs = []string{typ.Name()}
	ed.gui.crumbs.Text = ed.crumbs[0]

}

func (ed *Editor) initCursor() {

	ed.cursor = []int{0}
	ed.initCrumbs()
	ed.fields()

}

func (ed *Editor) updateCursorGraphic() {

	index := ed.cursor[len(ed.cursor)-1]
	display := []string{}
	for i, str := range ed.nav {

		num := i
		num++
		if num == 10 {
			num = 0
		}
		// numStr := strconv.Itoa(num)

		if i == index {
			display = append(display, "· "+str)
		} else {
			display = append(display, " "+str)
		}
	}

	ed.gui.nav.Items = display

}

func (ed *Editor) fields() {

	// typ := reflect.TypeOf(ed.obj)
	typ := ed.getParent()
	ed.Log("%v", typ)
	k := typ.NumField()
	fields := []string{}

	for i := 0; i < k; i++ {

		field := typ.Field(i)
		tag, ok := field.Tag.Lookup("nav")
		if ok {
			fields = append(fields, tag)
		} else {
			fields = append(fields, field.Name)
		}

	}

	ed.nav = fields

	ed.updateCursorGraphic()

}

func (ed *Editor) getParent() reflect.Type {
	x := reflect.TypeOf(ed.obj)

	if len(ed.cursor) == 1 {
		return x
	}

	field := x.Field(ed.cursor[0]).Type

	for i := 1; i < len(ed.cursor)-1; i++ {
		field = field.Field(ed.cursor[i]).Type
	}

	return field
}

func (ed *Editor) getType() reflect.Type {

	x := reflect.TypeOf(ed.obj)
	//field := x.Field(ed.cursor[0])
	field := x.Field(ed.cursor[0]).Type

	for i := 1; i < len(ed.cursor); i++ {
		if field.Kind() == reflect.Struct {
			field = field.Field(ed.cursor[i]).Type
		} else {
			field = field.Field(0).Type
		}
	}

	return field
}

func (ed *Editor) getName() string {

	x := reflect.TypeOf(ed.obj)
	//field := x.Field(ed.cursor[0])
	field := x.Field(ed.cursor[0])

	for i := 1; i < len(ed.cursor); i++ {
		if field.Type.Kind() == reflect.Struct {
			field = field.Type.Field(ed.cursor[i])
		} else {
			field = field.Type.Field(0)
		}
	}

	// TODO: Logic to take Tags over Names if available
	return field.Name

}

func (ed *Editor) getTag() reflect.StructField {
	x := reflect.TypeOf(ed.obj)
	var field reflect.StructField
	if len(ed.cursor) > 1 {
		field = x.FieldByIndex(ed.cursor)
		// field = x.Field(ed.cursor[len(ed.cursor)-1])
	} else {
		field = x.FieldByIndex(ed.cursor)
		// field = x.Field(ed.cursor[0])
	}

	// tag, ok := field.Tag.Lookup("nav")
	// return field.Type.Field(0).Name
	return field
}
