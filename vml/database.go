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

	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/sisatech/sherlock"
)

const (
	sqliteFK = "sqlite_fk"

	application = "app"
	directory   = "dir"
	none        = "none"
)

func init() {

	hook := func(conn *sqlite3.SQLiteConn) error {
		_, err := conn.Exec("PRAGMA foreign_keys = ON", nil)
		return err
	}

	driver := &sqlite3.SQLiteDriver{ConnectHook: hook}
	sql.Register(sqliteFK, driver)

}

func (repo *TinyRepo) database(path string) {

	var err error

	repo.db, err = sql.Open(sqliteFK, path)
	sherlock.Check(err)

	sherlock.Check(repo.db.Exec(`CREATE TABLE IF NOT EXISTS nodes (
			name TEXT,
			path TEXT,
			parent TEXT,
			type TEXT,
			PRIMARY KEY (path),
			FOREIGN KEY (parent) REFERENCES nodes(path)
		)`))

	sherlock.Check(repo.db.Exec(`CREATE TABLE IF NOT EXISTS apps (
			path TEXT,
			ref TEXT,
			date INT,
			extra INT,
			tag TEXT,
			app TEXT,
			files TEXT,
			icon TEXT,
			config TEXT,
			PRIMARY KEY (ref),
			FOREIGN KEY (path) REFERENCES nodes(path)
		)`))

	var n int
	row := repo.db.QueryRow(`SELECT COUNT(*) FROM nodes`)
	sherlock.Check(row.Scan(&n))

	if n == 0 {
		sherlock.Check(repo.db.Exec(`INSERT INTO nodes (name,
			path, parent, type) VALUES (?,?,?,?)`, "", "", "",
			directory))
	}

}
