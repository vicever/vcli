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
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/sisatech/sherlock"
)

func checksum(file io.Reader) string {

	hasher := sha256.New()
	sherlock.Check(io.Copy(hasher, file))

	return hex.EncodeToString(hasher.Sum(nil))

}

func reference(app, cfg, icon, files string, timestamp, extra int64) string {

	return checksum(bytes.NewBufferString(app + cfg + icon + files + strconv.Itoa(int(timestamp)) + strconv.Itoa(int(extra))))

}

func parent(path string) string {

	i := strings.LastIndex(path, "/")
	if i == -1 {
		return ""
	}

	return path[:i]

}

func split(path string) (string, string) {

	i := strings.LastIndex(path, "/")
	if i == -1 {
		return "", path
	}

	return path[:i], path[i+1:]

}

func errParentNodeIsApp(parent string) error {

	return fmt.Errorf("can't make child of directory node: %v", parent)

}

func errNodeIsDir(parent string) error {

	return fmt.Errorf("target is a directory: %v", parent)

}

func (repo *TinyRepo) nodeType(path string) string {

	var node string

	row := repo.db.QueryRow(`SELECT type FROM nodes WHERE path=?`, path)
	err := row.Scan(&node)
	sherlock.Assert(err, err == nil || err == sql.ErrNoRows)

	if err == sql.ErrNoRows {
		return none
	}

	return node

}
