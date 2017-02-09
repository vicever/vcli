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
package archiver

import (
	"io"
	"os"
)

// Archiver defines a common interface for different types of storage packages.
type Archiver interface {
	Put(input io.Reader) (string, error)
	Get(filename string) (io.ReadCloser, error)
	List() []os.FileInfo
	Delete(filename string) error
}
