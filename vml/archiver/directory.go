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
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sisatech/sherlock"
)

// TODO: change implementation to not require a temporary file to store the tar
// file.

type TarWriter struct {
	prefix string
	file   *os.File
	tar    *tar.Writer
}

// NewTarWriter returns an object that writes an entire directory recursively
// into a tar stream.
//
// CONTRACT: 'path' must point at an accessible directory.
// CONTRACT: 'path' must not be the root directory.
func NewTarWriter(path string) (*TarWriter, error) {

	var w *TarWriter
	var err error

	err = sherlock.Try(func() {

		path = strings.TrimSuffix(path, "/")

		w = &TarWriter{
			prefix: path,
		}

		w.file, err = ioutil.TempFile("", "")
		sherlock.Check(err)

		w.tar = tar.NewWriter(w.file)

		w.build()

	})

	return w, err

}

// Implemented to satisfy io.Reader interface.
func (w *TarWriter) Read(p []byte) (n int, err error) {

	return w.file.Read(p)

}

// Implemented to satisfy io.Closer interface.
func (w *TarWriter) Close() error {

	defer os.Remove(w.file.Name())
	return w.file.Close()

}

func (w *TarWriter) build() {

	var err error

	w.recurse(w.prefix, w.prefix)

	sherlock.Check(w.tar.Flush())

	sherlock.Check(w.tar.Close())

	sherlock.Check(w.file.Close())

	w.file, err = os.Open(w.file.Name())
	sherlock.Check(err)

}

func (w *TarWriter) recurse(path, prefix string) {

	info, err := os.Stat(path)
	sherlock.Check(err)

	if info.IsDir() {

		// add children to list
		children, err := ioutil.ReadDir(path)
		sherlock.Check(err)

		for _, child := range children {

			w.recurse(path+"/"+child.Name(), prefix)

		}

	} else {

		// setup next reader
		file, err := os.Open(path)
		sherlock.Check(err)

		hdr, err := tar.FileInfoHeader(info, "")
		sherlock.Check(err)

		hdr.Name = strings.Trim(strings.TrimPrefix(path, prefix), "/")

		sherlock.Check(w.tar.WriteHeader(hdr))

		sherlock.Check(io.Copy(w.tar, file))

	}

}

// ExpandTarStream takes a reader of a tar stream and writes a directory to the
// filesystem.
func ExpandTarStream(r io.Reader, path string) error {

	return sherlock.Try(func() {

		sherlock.Check(os.MkdirAll(path, 0777))

		tr := tar.NewReader(r)

		for hdr, err := tr.Next(); err == nil; hdr, err = tr.Next() {

			sherlock.Assert(err, err == nil || err == io.EOF)

			elems := strings.Split(hdr.Name, "/")
			if l := len(elems); l > 1 {

				parent := strings.Join(elems[:l-1], "/")

				sherlock.Check(os.MkdirAll(path+"/"+parent, 0777))

			}

			f, err := os.Create(path + "/" + hdr.Name)
			sherlock.Check(err)

			_, err = io.Copy(f, tr)
			f.Close()
			sherlock.Check(err)

		}

	})

}
