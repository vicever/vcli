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
package vml

const (
	insert = "insert"
	remove = "delete"
	modify = "modify"
)

var (
	summary Summary
)

// Summary is a summary of actions performed on the database.
type Summary []*ActionTuple

// ActionTuple represents a single operation performed on the database.
type ActionTuple struct {
	Address string
	Action  string
	Type    string
}

func (repo *TinyRepo) summaryReset() {

	summary = make([]*ActionTuple, 0)

}

func (repo *TinyRepo) summaryInsert(address, nodeType string) {

	summary = append(summary, &ActionTuple{
		Address: address,
		Action:  insert,
		Type:    nodeType,
	})

}

func (repo *TinyRepo) summaryModify(address, nodeType string) {

	if l := len(summary); l > 0 {

		last := summary[l-1]

		if (last.Address == address) && (last.Type == nodeType) {
			return
		}

	}

	summary = append(summary, &ActionTuple{
		Address: address,
		Action:  modify,
		Type:    nodeType,
	})

}

func (repo *TinyRepo) summaryRemove(address, nodeType string) {

	backspace := func() bool {

		l := len(summary)
		if l == 0 {
			return false
		}

		last := summary[l-1]
		if last.Address == address && last.Type == nodeType &&
			last.Action == modify {

			summary = summary[:l-1]
			return true

		}

		return false

	}

	for backspace() {
	}

	summary = append(summary, &ActionTuple{
		Address: address,
		Action:  modify,
		Type:    nodeType,
	})

}
