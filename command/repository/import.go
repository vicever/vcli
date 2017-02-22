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
package cmdrepo

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/command"
	"github.com/sisatech/vcli/compiler/converter"
	"github.com/sisatech/vcli/home"
	"github.com/sisatech/vcli/shared"
	"github.com/sisatech/vcli/vml"
)

type cmdImport struct {
	*kingpin.CmdClause
	src    string
	files  string
	icon   string
	addr   string
	tag    string
	config string
}

// New ...
func newImportCmd() *cmdImport {

	return &cmdImport{}

}

// Attach ...
func (cmd *cmdImport) Attach(parent command.Node) {

	cmd.CmdClause = parent.Command("import", shared.Catenate(`The import
		command takes the target Vorteil zip archive and stores it
		within the local repository.`))

	clause := cmd.Arg("source", shared.Catenate(`Path to the valid Vorteil
		zip archive or a plain ELF binary. If targeting a binary, a
		Vorteil build config file mest exist at <source>.vcfg.`))
	clause.Required()
	// clause.StringVar(&cmd.src)
	clause.ExistingFileVar(&cmd.src)

	clause = cmd.Arg("address", shared.Catenate(`Address within the local
		repository to store the imported app.`))
	clause.StringVar(&cmd.addr)

	flag := cmd.Flag("files", shared.Catenate(`Location of a folder to be
		cloned for the root directory of the app. The folder will be
		stored within the repository and tied to the app in its current
		state for version control purposes. This flag should only be
		provided when importing loose files, and will be ignored for
		zip archives.`))
	// flag.StringVar(&cmd.files)
	flag.ExistingDirVar(&cmd.files)

	flag = cmd.Flag("icon", shared.Catenate(`Location of a picture to be
		used as the icon for the app. This flag should only be
		provided when importing loose files, and will be ignored for
		zip archives.`))
	// flag.StringVar(&cmd.icon)
	flag.ExistingFileVar(&cmd.icon)
	flag.Hidden()

	flag = cmd.Flag("tag", shared.Catenate(`Tag the new version of the app
		during the import process with the provided name.`))
	flag.StringVar(&cmd.tag)

	flag = cmd.Flag("config", shared.Catenate(`Override the default config
		file associated with the app. Specifies path to desired config.`))
	flag.Short('c')
	flag.StringVar(&cmd.config)

	cmd.Action(cmd.action)

}

func (cmd *cmdImport) argValidation() error {
	// icon - PNG
	if cmd.icon != "" {
		f, err := os.Open(cmd.icon)
		if err != nil {
			return err
		}
		b := make([]byte, 512)
		f.Read(b)
		typ := http.DetectContentType(b)

		if typ != "image/png" {
			fmt.Printf("Detected filetype was: %v.\n", typ)
			return errors.New("icon must be a PNG file")
		}

	}

	// if cmd.src != "" {
	//
	// 	// Check if zip
	// 	f, err := os.Open(cmd.src)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	b := make([]byte, 512)
	// 	f.Read(b)
	// 	typ := http.DetectContentType(b)
	//
	// 	if typ != "application/zip" {
	// 		err := compiler.ValidateELF(cmd.src)
	// 		if err != nil {
	// 			return errors.New("source file is neither ELF nor ZIP file type")
	// 		}
	// 	}
	//
	// 	cf, err := ioutil.ReadFile(cmd.src + ".vcfg")
	// 	if err != nil {
	// 		return errors.New("could not open file " + cmd.src + ".vcfg")
	// 	}
	//
	// 	var conf shared.BuildConfig
	//
	// 	err = yaml.Unmarshal(cf, &conf)
	// 	if err != nil {
	// 		return errors.New("invalid or corrupt config file.")
	// 	}
	//
	// }

	// Validate Address
	addSpl := strings.Split(cmd.addr, "/")
	for _, x := range addSpl {
		err := cmd.valChars(x)
		if err != nil {
			return errors.New("illegal characters detected in provided address")
		}
	}

	// Validate tag
	if cmd.tag != "" {
		err := cmd.valChars(cmd.tag)
		if err != nil {
			return errors.New("illegal characters detected in provided tag")
		}
	}

	return nil
}

func (cmd *cmdImport) valChars(str string) error {
	for _, x := range str {
		if x == '.' || x == '-' || x == '_' {
			continue
		} else if x >= '0' && x <= '9' {
			continue
		} else if x >= 'a' && x <= 'z' {
			continue
		} else if x >= 'A' && x <= 'Z' {
			continue
		} else {
			return errors.New("illegal characters detected")
		}
	}

	return nil
}

func (cmd *cmdImport) action(ctx *kingpin.ParseContext) error {

	return sherlock.Try(func() {

		err := cmd.argValidation()
		if err != nil {
			sherlock.Check(err)
		}

		if cmd.addr == "" {
			x := strings.Split(cmd.src, "/")
			cmd.addr = x[len(x)-1]
		}

		if strings.HasPrefix(cmd.addr, shared.RepoPrefix) {
			cmd.addr = strings.TrimPrefix(cmd.addr, shared.RepoPrefix)
		}

		var config string
		if cmd.config == "" {
			config = cmd.src + ".vcfg"
		} else {
			config = cmd.config
		}

		// load input
		var in converter.Convertible

		if strings.HasPrefix(cmd.src, shared.RepoPrefix) {

			// load input from repository
			repo, err := vml.NewTinyRepo(home.Path(home.Repository))
			sherlock.Check(err)

			defer repo.Close()

			in, err = repo.Export(strings.TrimPrefix(cmd.src, shared.RepoPrefix), "")
			sherlock.Check(err)

		} else {

			// load input from filesystem
			if shared.IsELF(cmd.src) {

				in, err = converter.LoadLoose(cmd.src, config, cmd.icon, cmd.files)
				if err != nil {
					sherlock.Check(err)
				}

			} else if shared.IsZip(cmd.src) {

				in, err = converter.LoadZipFile(cmd.src)
				if err != nil {
					sherlock.Check(err)
				}

			} else {

				sherlock.Check(errors.New("target not a valid ELF or zip archive"))

			}

		}

		defer in.Close()

		mgr, err := vml.NewTinyRepo(home.Path(home.Repository))
		if err != nil {
			sherlock.Check(err)
		}

		hash, summary, err := mgr.Import(cmd.addr, in.App(), in.Config(), in.Icon(), in.FilesTar())
		if err != nil {
			sherlock.Check(err)
		}

		Summary(summary, false)

		fmt.Println("Imported application with reference id: " + hash)

	})

}

func Summary(summary vml.Summary, plain bool) {
	// table := tablewriter.NewWriter(os.Stdout)
	// if plain {
	// 	table.SetBorder(false)
	// 	table.SetHeaderLine(false)
	// 	table.SetCenterSeparator("")
	// 	table.SetColumnSeparator("")
	// } else {
	// 	table.SetHeader([]string{"ACTION", "NODE", "ADDRESS"})
	// }
	// operations := writer.Summary()
	// for _, x := range operations {
	// 	table.Append([]string{
	// 		actionTypes[x.Action],
	// 		nodeTypes[x.Type],
	// 		x.Address,
	// 	})
	// }
	// table.Render()
	// TODO: nice wrapping for >80 character rows.

	var vals [][]string
	vals = append(vals, []string{"Action", "Node", "Address"})
	for _, x := range summary {
		var row []string
		row = append(row, x.Action)
		row = append(row, x.Type)
		row = append(row, x.Address)

		vals = append(vals, row)
	}

	shared.PrettyLeftTable(vals)

}

// TODO: import file to local repository
// TODO: import support loose files and zip archives alike
// TODO: import icon file
// TODO: import auxiliary files
