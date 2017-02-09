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
package converter

import (
	"archive/tar"
	"archive/zip"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sisatech/sherlock"
)

var (
	ErrInvalidZip = errors.New("invalid zip archive")
)

type Zip struct {
	tmp    string
	out    *os.File
	app    io.ReadCloser
	cfg    io.ReadCloser
	icon   io.ReadCloser
	files  io.ReadCloser
	zipper *zip.ReadCloser
}

func (out *Zip) App() io.Reader {

	return out.app

}

func (out *Zip) Config() io.Reader {

	return out.cfg

}

func (out *Zip) Icon() io.Reader {

	return out.icon

}

func (out *Zip) FilesTar() io.Reader {

	return out.files

}

func (out *Zip) Close() error {

	return sherlock.Try(func() {

		sherlock.Check(out.app.Close())

		sherlock.Check(out.cfg.Close())

		if out.icon != nil {
			sherlock.Check(out.icon.Close())
		}

		if out.files != nil {
			sherlock.Check(out.files.Close())
		}

		if out.tmp != "" {
			sherlock.Check(os.Remove(out.tmp))
		}

		sherlock.Check(out.zipper.Close())

	})

}

func LoadZipFile(path string) (*Zip, error) {

	var err error
	out := new(Zip)

	err = sherlock.Try(func() {

		out.zipper, err = zip.OpenReader(path)
		sherlock.Check(err)

		sherlock.Assert(ErrInvalidZip, len(out.zipper.File) > 1)

		// app
		out.app, err = out.zipper.File[0].Open()
		sherlock.Check(err)

		// config
		out.cfg, err = out.zipper.File[1].Open()
		sherlock.Check(err)

		// more
		l := len(out.zipper.File)
		if l > 2 {

			// analyse name to determine if an icon was included
			index := 2
			var iconExists bool
			name := out.zipper.File[2].Name
			if !strings.Contains(name, "/") {
				iconExists = true
				index++
			}

			// load icon if it exists
			if iconExists {
				out.icon, err = out.zipper.File[2].Open()
				sherlock.Check(err)
			}

			// load filesystem if it exists
			tmp, err := ioutil.TempFile("", "")
			sherlock.Check(err)
			defer tmp.Close()

			out.tmp = tmp.Name()

			fi, err := os.Stat(tmp.Name())
			sherlock.Check(err)
			hdr, err := tar.FileInfoHeader(fi, "")
			sherlock.Check(err)

			tarrer := tar.NewWriter(tmp)

			for i := index; i < l; i++ {

				hdr.Name = strings.TrimPrefix(out.zipper.File[i].Name, "fs/")
				hdr.Size = out.zipper.File[i].FileInfo().Size()

				sherlock.Check(tarrer.WriteHeader(hdr))

				w, err := out.zipper.File[i].Open()
				sherlock.Check(err)

				sherlock.Check(io.Copy(tarrer, w))

			}

			sherlock.Check(tarrer.Close())

			sherlock.Check(tmp.Close())
			out.files, err = os.Open(tmp.Name())
			sherlock.Check(err)

		}

	})

	return out, err

}

func ExportZipFile(in Convertible, path string) (*os.File, error) {

	var err error
	var f *os.File

	err = sherlock.Try(func() {

		defer in.Close()

		if path == "" {
			f, err = ioutil.TempFile("", "")
			sherlock.Check(err)
			defer f.Close()
		} else {
			f, err = os.Create(path)
			sherlock.Check(err)
			defer f.Close()
		}

		zipper := zip.NewWriter(f)

		// app
		w, err := zipper.Create("app")
		sherlock.Check(err)

		sherlock.Check(io.Copy(w, in.App()))

		// config
		w, err = zipper.Create("app.vcfg")
		sherlock.Check(err)
		sherlock.Check(io.Copy(w, in.Config()))

		// icon
		if in.Icon() != nil {
			w, err = zipper.Create("app.png")
			sherlock.Check(err)
			sherlock.Check(io.Copy(w, in.Icon()))
		}

		// files
		if in.FilesTar() != nil {

			tarrer := tar.NewReader(in.FilesTar())

			for {

				hdr, err := tarrer.Next()
				if err != nil && err == io.EOF {
					break
				}
				sherlock.Check(err)

				w, err = zipper.Create("fs/" + hdr.Name)
				sherlock.Check(err)
				sherlock.Check(io.Copy(w, tarrer))

			}

		}

		sherlock.Check(zipper.Close())

	})

	if err == nil {

		f, err = os.Open(f.Name())

	}

	return f, err

}
