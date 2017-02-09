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
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sisatech/vcli/compiler/fs/ext2"
)

const (
	clean = false

	filesystem = "filesystem"
	mountpoint = "mountpoint"
	dir        = "files"
	inodes     = uint32(4096)
	sectors    = uint32(1048576) // 256 MiB
)

var (
	files = map[string]int{
		"alpha":   8,
		"bravo":   16,
		"charlie": 256,
		"delta":   512,
		"echo":    65536,
		"foxtrot": 67584,
	}
)

func main() {

	err := generateFiles()
	fail(err)

	// defer cleanup()

	// generate filesystem
	err = generateFilesystem()
	fail(err)
	defer os.Remove(filesystem)

	// run fsck
	fsck := exec.Command("e2fsck", "-f", "-n", "-v", filesystem)
	out, err := fsck.StdoutPipe()
	fail(err)
	go io.Copy(os.Stdout, out)
	err = fsck.Run()
	if err != nil {
		fail(fmt.Errorf("fsck error: %v", err))
	}

	fmt.Println("fsck PASSED")

	// mount filesystem
	err = os.MkdirAll(mountpoint, 0777)
	fail(err)
	defer os.Remove(mountpoint)

	mount := exec.Command("mount", filesystem, mountpoint)
	out, err = mount.StdoutPipe()
	fail(err)
	go io.Copy(os.Stdout, out)
	err = mount.Run()
	if err != nil {
		fail(fmt.Errorf("mount error: %v", err))
	}

	unmount := exec.Command("umount", mountpoint)
	defer unmount.Run()

	// compare files
	err = compare()
	fail(err)

}

func compare() error {

	for k := range files {

		orig, err := os.Open(dir + "/" + k)
		if err != nil {
			return err
		}

		defer orig.Close()

		test, err := os.Open(mountpoint + "/" + k)
		if err != nil {
			return err
		}

		defer test.Close()

		// compare contents of files
		a := make([]byte, 1024)
		b := make([]byte, 1024)
		for {

			n1, err1 := orig.Read(a)
			if err1 != nil && err1 != io.EOF {
				return err1
			}

			n2, err2 := test.Read(b)
			if err2 != nil && err2 != io.EOF {
				return err2
			}

			if (err1 == nil && err2 != nil) || (err1 != nil && err2 == nil) {
				return fmt.Errorf("file %s: not the same length", k)
			}

			if n1 != n2 {
				return fmt.Errorf("file %s: not the same length", k)
			}

			if n1 == 0 {
				break
			}

			if bytes.Compare(a, b) != 0 {
				return fmt.Errorf("file %s: not the same content", k)
			}

		}

		orig.Close()
		test.Close()

	}

	return nil

}

func generateFilesystem() error {

	ins, err := ext2.Compile(dir, sectors, inodes)
	if err != nil {
		return err
	}

	fs, err := os.Create(filesystem)
	if err != nil {
		return err
	}
	defer fs.Close()

	for ins.Next() {

		_, err = fs.Seek(ins.Offset(), 0)
		if err != nil {
			return err
		}

		_, err = io.CopyN(fs, ins.Data(), ins.Length())
		if err != nil {
			return err
		}

	}

	// pad to end
	// TODO

	return nil

}

func cleanup() {

	for k := range files {
		defer os.Remove(dir + "/" + k)
	}

	os.RemoveAll(dir)

}

func fail(err error) {

	if err != nil {
		unmount := exec.Command("umount", mountpoint)
		unmount.Run()
		if clean {
			os.Remove(filesystem)
		}
		// cleanup()
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

}

func generateFiles() error {

	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return err
	}

	for k, v := range files {

		err = generateFile(dir+"/"+k, v)
		if err != nil {
			return err
		}

	}

	return nil

}

func generateFile(name string, size int) error {

	f, err := os.Create(name)
	if err != nil {
		return err
	}

	defer f.Close()

	for i := 0; i < size; i++ {

		slice := bytes.Repeat([]byte{uint8(i % 256)}, 1024)

		_, err = f.Write(slice)
		if err != nil {
			return err
		}

	}

	return nil

}
