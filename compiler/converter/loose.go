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
	"io"
	"os"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/vml/archiver"
)

type Loose struct {
	app   io.ReadCloser
	cfg   io.ReadCloser
	icon  io.ReadCloser
	files io.ReadCloser
}

func (out *Loose) App() io.Reader {

	return out.app

}

func (out *Loose) Config() io.Reader {

	return out.cfg

}

func (out *Loose) Icon() io.Reader {

	return out.icon

}

func (out *Loose) FilesTar() io.Reader {

	return out.files

}

func (out *Loose) Close() error {

	return sherlock.Try(func() {

		sherlock.Check(out.app.Close())

		sherlock.Check(out.cfg.Close())

		if out.icon != nil {
			sherlock.Check(out.icon.Close())
		}

		if out.files != nil {
			sherlock.Check(out.files.Close())
		}

	})

}

func LoadLoose(app, cfg, icon, files string) (*Loose, error) {

	loose := new(Loose)

	err := sherlock.Try(func() {

		var err error

		loose.app, err = os.Open(app)
		sherlock.Check(err)

		loose.cfg, err = os.Open(cfg)
		sherlock.Check(err)

		if icon != "" {
			loose.icon, err = os.Open(icon)
			sherlock.Check(err)
		}

		if files != "" {
			loose.files, err = archiver.NewTarWriter(files)
			sherlock.Check(err)
		}

	})

	return loose, err

}

func ExportLoose(in Convertible, path string) error {

	return sherlock.Try(func() {

		defer in.Close()

		if path == "" {
			path = "."
		}
		sherlock.Check(os.MkdirAll(path, 0777))

		// app
		f, err := os.Create(path + "/app")
		sherlock.Check(err)
		defer f.Close()
		sherlock.Check(io.Copy(f, in.App()))

		// config
		f, err = os.Create(path + "/app.vcfg")
		sherlock.Check(err)
		defer f.Close()
		sherlock.Check(io.Copy(f, in.Config()))

		// icon
		if in.Icon() != nil {
			f, err = os.Create(path + "/icon.png")
			sherlock.Check(err)
			defer f.Close()
			sherlock.Check(io.Copy(f, in.Icon()))
		}
		// files
		if in.FilesTar() != nil {
			archiver.ExpandTarStream(in.FilesTar(), path+"/fs")
		} else {
			sherlock.Check(os.Mkdir(path+"/fs", 0777))
		}

	})

}
