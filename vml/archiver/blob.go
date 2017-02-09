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
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"sync"

	"github.com/sisatech/sherlock"
)

const (
	nilChecksum = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// BlobArchiver implements Archiver and stores files by computing their SHA256
// checksums to prevent blob duplication.
type BlobArchiver struct {
	lock     sync.RWMutex
	path     string
	contents contents
	compress bool
	level    int
}

// New returns a new BlobArchiver using the provided path as the base folder to
// store blobs.
//
// CONTRACT: if 'path' exists, it must not contain a directory.
// CONTRACT: if 'path' exists, it must be readable, writable, and executable by
// 	the process, and all files within must also be readable by the process.
func New(path string) (*BlobArchiver, error) {

	archive := new(BlobArchiver)
	archive.path = path

	return archive, sherlock.Try(archive.init)

}

// NewCompressed returns a new BlobArchiver using the provided path as the base
// folder to store blobs, using gz compression at the default compression level.
//
// CONTRACT: if 'path' exists, it must not contain a directory.
// CONTRACT: if 'path' exists, it must be readable, writable, and executable by
// 	the process, and all files within must also be readable by the process.
// CONTRACT: if 'path' exists, any files within must be gz compressed.
func NewCompressed(path string) (*BlobArchiver, error) {

	return NewCompressedLevel(path, -1)

}

// NewCompressed returns a new BlobArchiver using the provided path as the base
// folder to store blobs, using gz compression at the given compression level.
// -1 is default compression, 0 is no compression, 1-9 are varying levels of
// compression trading speed (0) for filesize (9).
//
// CONTRACT: if 'path' exists, it must not contain a directory.
// CONTRACT: if 'path' exists, it must be readable, writable, and executable by
// 	the process, and all files within must also be readable by the process.
// CONTRACT: if 'path' exists, any files within must be gz compressed.
func NewCompressedLevel(path string, level int) (*BlobArchiver, error) {

	archive, err := New(path)
	if err != nil {
		return nil, err
	}

	archive.compress = true
	archive.level = level

	return archive, nil

}

func (archive *BlobArchiver) filepath(filename string) string {

	return archive.path + "/" + filename

}

func (archive *BlobArchiver) init() {

	var err error

	sherlock.Assert(errors.New("empty path string"), archive.path != "")
	sherlock.Assert(errors.New("can't use root directory"), archive.path != "/")

	// ensure folder exists
	sherlock.Check(os.MkdirAll(archive.path, 0777))

	// load file info
	archive.contents, err = ioutil.ReadDir(archive.path)
	sherlock.Check(err)

}

// Put adds the data stored within 'in' as a new file to the archive.
func (archive *BlobArchiver) Put(in io.Reader) (string, error) {

	if in == nil {
		return nilChecksum, nil
	}

	var checksum string

	err := sherlock.Try(func() {

		var target io.WriteCloser

		// write input to temporary file
		tmp, err := ioutil.TempFile("", "")
		sherlock.Check(err)
		defer os.Remove(tmp.Name())
		defer tmp.Close()

		target = tmp

		// apply gz compression if enabled
		if archive.compress {

			target, err = gzip.NewWriterLevel(target, archive.level)
			sherlock.Check(err)

		}

		// compute SHA256 checksum of input
		hasher := sha256.New()

		w := io.MultiWriter(hasher, target)

		sherlock.Check(io.Copy(w, in))

		checksum = hex.EncodeToString(hasher.Sum(nil))

		// take a write lock on the archiver
		archive.lock.Lock()
		defer archive.lock.Unlock()

		// lookup index within info list
		index := sort.Search(len(archive.contents), func(i int) bool {

			name := archive.contents[i].Name()

			return checksum <= name

		})

		// return if value exists within contentss
		if index != len(archive.contents) &&
			archive.contents[index].Name() == checksum {
			return
		}

		// move file into directory
		filename := archive.filepath(checksum)
		err = tmp.Close()
		sherlock.Check(err)
		sherlock.Check(os.Rename(tmp.Name(), filename))

		// add file info to contents
		info, err := os.Stat(filename)
		sherlock.Check(err)
		archive.contents = append(archive.contents, info)

		// sort contents
		sort.Sort(&archive.contents)

	})

	return checksum, err

}

// Get returns a io.ReadCloser for the desired file.
func (archive *BlobArchiver) Get(filename string) (io.ReadCloser, error) {

	var err error
	var ret io.ReadCloser

	if filename == nilChecksum {
		return nil, nil
	}

	err = sherlock.Try(func() {

		ret, err = os.Open(archive.filepath(filename))
		sherlock.Check(err)

		// apply gz decompression on the way out if required
		if archive.compress {

			ret, err = gzip.NewReader(ret)
			sherlock.Check(err)

		}

	})

	return ret, err

}

// List returns the complete list of files stored within the BlobArchiver.
func (archive *BlobArchiver) List() []os.FileInfo {

	archive.lock.RLock()
	defer archive.lock.RUnlock()

	return archive.contents

}

// Delete removes a file from the archive.
func (archive *BlobArchiver) Delete(filename string) error {

	return sherlock.Try(func() {
		if filename == nilChecksum {
			return
		}

		archive.lock.Lock()
		defer archive.lock.Unlock()

		// remove file from disk
		sherlock.Check(os.Remove(archive.filepath(filename)))

		// lookup index within info list
		index := sort.Search(len(archive.contents), func(i int) bool {

			return filename == archive.contents[i].Name()

		})

		if index == len(archive.contents) {
			return
		}

		archive.contents = append(archive.contents[:index],
			archive.contents[index+1:]...)

	})

}
