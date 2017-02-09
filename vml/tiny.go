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

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/vml/archiver"
)

var (
	ErrNotFound = errors.New("target not found")
)

// TinyRepo provides a very simple and minimal package for version controlling
// local Vorteil applications. TinyRepo is not threadsafe.
type TinyRepo struct {
	archive archiver.Archiver
	db      *sql.DB
}

// NewTinyRepo returns a new TinyRepo based on the given path.
func NewTinyRepo(path string) (*TinyRepo, error) {

	var err error
	var repo *TinyRepo

	err = sherlock.Try(func() {

		repo = new(TinyRepo)

		// initialize archive
		repo.archive, err = archiver.New(path)
		sherlock.Check(err)

		// TODO NewCompressed. Currently bugged.

		// initialize database
		repo.database(path + "/tiny.db")

	})

	return repo, err

}

// Close cleans up all resources in use by TinyRepo
func (repo *TinyRepo) Close() error {

	return sherlock.Try(func() {

		sherlock.Check(repo.db.Close())

	})

}

type ExportOutput struct {
	app    io.ReadCloser
	cfg    io.ReadCloser
	icon   io.ReadCloser
	files  io.ReadCloser
	closed bool
}

func (out *ExportOutput) App() io.Reader {

	return out.app

}

func (out *ExportOutput) Config() io.Reader {

	return out.cfg

}

func (out *ExportOutput) Icon() io.Reader {

	return out.icon

}

func (out *ExportOutput) FilesTar() io.Reader {

	return out.files

}

func (out *ExportOutput) Close() error {

	return sherlock.Try(func() {

		if out.closed {
			return
		}

		out.closed = true

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

func (repo *TinyRepo) Export(path, ref string) (*ExportOutput, error) {

	var err error
	out := &ExportOutput{}

	err = sherlock.Try(func() {

		// find target to export
		hash := repo.deref(path, ref)

		sherlock.Assert(ErrNotFound, hash != "")

		row := repo.db.QueryRow(`SELECT app, files, icon, config
			FROM apps WHERE path=$1 AND ref=$2`, path, hash)

		var appHash, cfgHash, iconHash, filesHash string
		sherlock.Check(row.Scan(&appHash, &filesHash, &iconHash, &cfgHash))

		out.app, err = repo.archive.Get(appHash)
		sherlock.Check(err)

		out.cfg, err = repo.archive.Get(cfgHash)
		sherlock.Check(err)

		out.icon, err = repo.archive.Get(iconHash)
		sherlock.Check(err)

		out.files, err = repo.archive.Get(filesHash)
		sherlock.Check(err)

	})

	return out, err

}

// Import adds the given application elements to the TinyRepo. The application
// elements are io.ReadClosers. Importantly, files is a tar produced by running
// archiver.NewTarWriter on the files directory.
func (repo *TinyRepo) Import(path string, app, cfg, icon,
	files io.Reader) (string, Summary, error) {

	repo.summaryReset()

	var err error
	var ref string

	err = sherlock.Try(func() {

		appID, err := repo.archive.Put(app)
		sherlock.Check(err)

		cfgID, err := repo.archive.Put(cfg)
		sherlock.Check(err)

		iconID, err := repo.archive.Put(icon)
		sherlock.Check(err)

		filesID, err := repo.archive.Put(files)
		sherlock.Check(err)

		ref = repo.addApp(path, appID, cfgID, iconID, filesID)

	})

	return ref, summary, err

}

func (repo *TinyRepo) addAppNode(path string) {

	// check if target exists
	node := repo.nodeType(path)
	switch node {

	case application:
		return

	case directory:
		sherlock.Throw(errNodeIsDir(path))

	case none:
		// check if parents exist
		parent, name := split(path)
		repo.addDirNode(parent)

		// create app node
		sherlock.Check(repo.db.Exec(`INSERT INTO nodes (name, path, parent, type)
			VALUES (?,?,?,?)`, name, path, parent, application))

		repo.summaryInsert(path, application)

	default:
		panic(nil)

	}

}

func (repo *TinyRepo) addDirNode(path string) {

	// check if target exists
	node := repo.nodeType(path)
	switch node {

	case application:
		sherlock.Throw(errParentNodeIsApp(path))

	case directory:
		return

	case none:
		// check if parents exist
		parent, name := split(path)
		repo.addDirNode(parent)

		// create app node
		sherlock.Check(repo.db.Exec(`INSERT INTO nodes (name, path, parent, type)
			VALUES (?,?,?,?)`, name, path, parent, directory))

		repo.summaryInsert(path, directory)

	default:
		panic(nil)

	}

}

func (repo *TinyRepo) addApp(path, app, cfg, icon, files string) string {

	repo.addAppNode(path)

	t := time.Now()
	timestamp := t.Unix()
	extra := t.UnixNano()

	ref := reference(app, cfg, icon, files, timestamp, extra)[:8]

	sherlock.Check(repo.db.Exec(`INSERT INTO apps (path, ref, date, extra, tag, app, files, icon, config)
		VALUES (?,?,?,?,?,?,?,?,?)`, path, ref, timestamp, extra, "", app, files, icon, cfg))

	repo.summaryModify(path, application)

	return ref

}

func (repo *TinyRepo) deref(path, ref string) string {

	// if ref not provided, get most recent
	if ref == "" {

		row := repo.db.QueryRow(`SELECT ref FROM apps
			WHERE path=$1 ORDER BY date DESC LIMIT 1`, path)

		var tmp string
		err := row.Scan(&tmp)

		sherlock.Assert(ErrNotFound, err != sql.ErrNoRows)
		sherlock.Assert(err, err == nil)

		return tmp

	}

	// ref was provided - lookup
	var n int

	rows, err := repo.db.Query(`SELECT ref, tag FROM apps
		WHERE path=$1 AND (ref=$2 OR tag=$2)`, path, ref)
	sherlock.Check(err)
	defer rows.Close()

	var tmp string

	for rows.Next() {

		n++

		var tmpRef, tmpTag string
		sherlock.Check(rows.Scan(&tmpRef, &tmpTag))

		if tmpRef == ref {
			return ref
		}

		tmp = tmpRef

	}

	return tmp

}

// Tag sets the tag value of an app to the given value. The app version is
// dereferenced from the ref argument, so it can apply to a version hash or an
// existing tag equally. It applied to an existing tag, it will overwrite it.
func (repo *TinyRepo) Tag(path, ref, tag string) (Summary, error) {

	repo.summaryReset()

	err := sherlock.Try(func() {

		// check if tag is in use
		hash := repo.deref(path, tag)
		sherlock.Assert(errors.New("tag already in use"), hash == "")

		// find target to tag
		hash = repo.deref(path, ref)
		sherlock.Assert(ErrNotFound, hash != "")

		sherlock.Check(repo.db.Exec(`UPDATE apps SET tag=$1
			WHERE path=$2 AND ref=$3`, tag, path, hash))

		repo.summaryModify(path, application)

	})

	return summary, err

}

func (repo *TinyRepo) Untag(path, ref string) (Summary, error) {

	repo.summaryReset()

	err := sherlock.Try(func() {

		// find target to tag
		hash := repo.deref(path, ref)
		sherlock.Assert(ErrNotFound, hash != "")

		sherlock.Check(repo.db.Exec(`UPDATE apps SET tag=$1
			WHERE path=$2 AND ref=$3`, "", path, hash))

		repo.summaryModify(path, application)

	})

	return summary, err

}

func (repo *TinyRepo) Delete(path, ref string) (Summary, error) {

	repo.summaryReset()

	err := sherlock.Try(func() {

		// check if target exists
		ref = repo.deref(path, ref)
		sherlock.Assert(ErrNotFound, ref != "")

		// delete target
		repo.deleteApp(path, ref)

		// check if target was last and cleanup
		var k int
		row := repo.db.QueryRow(`SELECT COUNT(*) FROM apps WHERE path=$1`, path)
		sherlock.Check(row.Scan(&k))

		if k == 0 {
			repo.deleteAppNode(path)
		}

	})
	return summary, err

}

func (repo *TinyRepo) deleteApp(path, ref string) {

	row := repo.db.QueryRow(`SELECT app, files, icon, config FROM apps
		WHERE path=$1 AND ref=$2`, path, ref)

	var app, files, icon, config string
	sherlock.Check(row.Scan(&app, &files, &icon, &config))

	sherlock.Check(repo.db.Exec(`DELETE FROM apps WHERE path=$1 AND ref=$2`,
		path, ref))

	repo.summaryModify(path, application)

	// cleanup archive resources if possible
	var k int

	fn := func(hash string) {

		row = repo.db.QueryRow(`SELECT COUNT(*) FROM apps
			WHERE app=$1 OR files=$1 OR icon=$1 OR config=$1`, hash)
		sherlock.Check(row.Scan(&k))

		if k == 0 {
			sherlock.Check(repo.archive.Delete(hash))
		}

	}

	fn(app)
	fn(files)
	fn(icon)
	fn(config)

}

func (repo *TinyRepo) deleteAppNode(path string) {

	// delete app node
	sherlock.Check(repo.db.Exec(`DELETE FROM nodes WHERE path=$1`, path))

	// add to summary
	repo.summaryRemove(path, application)

	// recursively delete parent, except for root node
	parent := parent(path)
	repo.deleteDirNode(parent)

}

func (repo *TinyRepo) deleteDirNode(path string) {

	if path == "" {
		return
	}

	// delete app node
	sherlock.Check(repo.db.Exec(`DELETE FROM nodes WHERE path=$1`, path))

	// add to summary
	repo.summaryRemove(path, directory)

	// recursively delete parent, except for root node
	parent := parent(path)
	repo.deleteDirNode(parent)

}

func (repo *TinyRepo) DeleteAll(path string) (Summary, error) {

	repo.summaryReset()

	err := sherlock.Try(func() {

		// lookup each version
		rows, err := repo.db.Query(`SELECT ref FROM apps WHERE path=$1`,
			path)
		sherlock.Check(err)
		defer rows.Close()

		var refs []string
		for rows.Next() {

			var tmp string
			sherlock.Check(rows.Scan(&tmp))
			refs = append(refs, tmp)

		}

		sherlock.Check(rows.Close())

		sherlock.Assert(ErrNotFound, len(refs) > 0)

		// delete each version
		for _, ref := range refs {
			repo.deleteApp(path, ref)
		}

		// delete node
		repo.deleteAppNode(path)

	})

	return summary, err

}

func (repo *TinyRepo) NodeType(path string) (string, error) {

	var ntype = none

	err := sherlock.Try(func() {

		row := repo.db.QueryRow(`SELECT type FROM nodes
			WHERE path=$1`, path)

		sherlock.Check(row.Scan(&ntype))

	})

	if err != nil && err == sql.ErrNoRows {
		err = fmt.Errorf("node not found at path: %v", path)
	}

	return ntype, err

}

type DirectoryList []*DirectoryEntry

type DirectoryEntry struct {
	Name string
	Type string
}

func (repo *TinyRepo) ListDir(path string) (DirectoryList, error) {

	var list DirectoryList

	err := sherlock.Try(func() {

		rows, err := repo.db.Query(`SELECT name, type FROM nodes
			WHERE parent=$1 ORDER BY name ASC`, path)
		sherlock.Check(err)
		defer rows.Close()

		for rows.Next() {

			entry := new(DirectoryEntry)
			sherlock.Check(rows.Scan(&entry.Name, &entry.Type))
			list = append(list, entry)

		}

	})

	if path == "" {
		list = list[1:]
	}

	return list, err

}

type AppList []*AppEntry

type AppEntry struct {
	Uploaded  int64
	Version   string
	Alias     string
	Reference string
}

func (repo *TinyRepo) ListApp(path string) (AppList, error) {

	var list AppList

	err := sherlock.Try(func() {

		rows, err := repo.db.Query(`SELECT date, ref, tag FROM apps
			WHERE path=$1 ORDER BY date DESC`, path)
		sherlock.Check(err)
		defer rows.Close()

		for rows.Next() {

			entry := new(AppEntry)
			sherlock.Check(rows.Scan(&entry.Uploaded, &entry.Version,
				&entry.Alias))
			entry.Reference = entry.Version
			if entry.Alias != "" {
				entry.Reference = entry.Alias
			}
			list = append(list, entry)

		}

	})

	return list, err

}
