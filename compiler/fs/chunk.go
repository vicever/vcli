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
package fs

import "io"

type Chunker struct {
	ins   Instructions
	seek  int64
	empty bool
	rdr   io.Reader
	first bool
}

func NewChunker(ins Instructions) *Chunker {

	return &Chunker{
		ins:   ins,
		first: true,
	}

}

func (chunk *Chunker) Next(n int64) bool {

	start := chunk.seek
	end := start + n

	// check if first time
	if chunk.first {

		if !chunk.ins.Next() {
			return false
		}

		chunk.first = false

	}

	// check if current instruction has overlap
	off := chunk.ins.Offset()
	l := chunk.ins.Length()

	// overflow
	if off < start {

		if off+l > start {
			chunk.empty = false
			return true
		}

		// advance to point
		// TODO

		return false

	}

	// perfect align
	if off == start {
		chunk.empty = false
		return true
	}

	// intersection
	if off < end {
		chunk.empty = false
		return true
	}

	// no intersection
	if off >= end {
		chunk.empty = true
		return true
	}

	return false

}

func (chunk *Chunker) Empty() bool {

	return chunk.empty

}

func (chunk *Chunker) Data() io.Reader {

}
