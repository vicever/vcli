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
package compiler

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"strings"

	"github.com/sisatech/targz"
	"github.com/sisatech/vcli/compiler/rawsparse"
	"github.com/sisatech/vcli/home"
)

// BuildSparseVMDK returns the name of a compiled .vmdk disk image within a
// temporary folder. The caller should move the file to a non-temporary location.
func BuildRawSparse(binary, config, files, kernel, destination string, debug bool) (string, error) {

	err := FullValidation(binary, config, files, "", kernel)
	if err != nil {
		return "", err
	}

	// TODO: implement
	env, err := rawsparse.NewEnvironment(home.Path(home.Kernel))
	if err != nil {
		return "", err
	}

	env.GrantWriteAccess()
	env.LogToStdout()

	f, err := env.BuildSparse(&rawsparse.BuildArgs{
		Binary:      binary,
		Config:      config,
		Files:       files,
		Kernel:      kernel,
		Debug:       debug,
		Destination: destination,
	})
	if err != nil {
		return "", err
	}

	err = f.Close()
	if err != nil {
		return f.Name(), err
	}

	err = targz.ArchiveFile(f.Name(), destination)
	if err != nil {
		return "", err
	}

	// fmt.Printf("f.Name() = %v\nDestination = %v\n", f.Name(), destination)
	os.Remove(f.Name())
	return destination + ".tar.gz", nil

}

func gzipit(source, target string) error {
	reader, err := os.Open(source)
	if err != nil {
		return err
	}
	// target = strings.TrimSuffix(target, ".img")
	writer, err := os.Create(target)
	if err != nil {
		return err
	}
	defer writer.Close()

	archiver := gzip.NewWriter(writer)
	archiver.Name = target
	defer archiver.Close()

	_, err = io.Copy(archiver, reader)

	return err

}

func compress(source, output string) error {

	// Add .tar.gz suffic to output if not there already ...
	output = strings.TrimSuffix(output, ".tar.gz")
	output += ".tar.gz"

	// fmt.Println(output)

	// Check source exists ...
	if _, err := os.Stat(source); os.IsNotExist(err) {
		return err
	}

	// Open it ...
	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()

	// Stat the file for later use ...
	stat, err := os.Stat(f.Name())
	if err != nil {
		return err
	}

	// Create the archive file
	a, err := os.Create(output)
	if err != nil {
		return err
	}

	// Create gzip and tar writers ...
	g := gzip.NewWriter(a)
	t := tar.NewWriter(g)
	defer g.Close()
	defer t.Close()

	// Create tar header for file ...
	th := new(tar.Header)
	th.Name = source
	th.Size = stat.Size()
	th.Mode = int64(stat.Mode())
	th.ModTime = stat.ModTime()

	// Add header to archive ...
	err = t.WriteHeader(th)
	if err != nil {
		return err
	}

	// Write file to archive ...
	_, err = io.Copy(t, f)
	if err != nil {
		return err
	}

	return nil
}
