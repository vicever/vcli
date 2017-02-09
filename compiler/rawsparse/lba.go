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
package rawsparse

import (
	"fmt"
	"os"
)

type content struct {
	file       *os.File
	LBAs       uint64
	reserved   offsets
	config     offsets
	kernel     offsets
	trampoline offsets
	app        offsets
	files      offsets
	backup     offsets
}

func (build *builder) calculateLBAs() error {

	// reserved
	build.content.reserved.first = 0
	build.content.reserved.length = 34
	build.content.reserved.last = build.content.reserved.first +
		build.content.reserved.length - 1

	// config
	build.content.config.first = build.content.reserved.last + 1
	build.content.config.length = 32
	build.content.config.last = build.content.config.first +
		build.content.config.length - 1

	// kernel
	debug := "PROD"
	if build.args.Debug {
		debug = "DEBUG"
	}

	kernel := fmt.Sprintf("vkernel-%s-%s.img", debug, build.args.Kernel)
	path := fmt.Sprintf("%s/%s", build.env.path, kernel)

	info, err := os.Stat(path)
	if err != nil {

		if os.IsNotExist(err) {

			// TODO
			// try to download kernel
			// build.Log("%s not found locally. Trying to download.", kernel)
			// err := DownloadVorteilFile(kernel, "vkernel")
			if err != nil {
				return err
			}

			info, err = os.Stat(path)
			if err != nil {
				return err
			}

		} else {
			return err
		}

	}

	build.content.kernel.first = build.content.config.last + 1
	build.content.kernel.length = ceiling(uint64(info.Size()), SectorSize)
	build.content.kernel.last = build.content.kernel.first +
		build.content.kernel.length - 1

	// trampoline
	path = fmt.Sprintf("%s/%s", build.env.path, "vtramp.img")

	info, err = os.Stat(path)
	if err != nil {

		if os.IsNotExist(err) {

			// TODO
			// try to download trampoline
			// build.Log("%s not found locally. Trying to download.", kernel)
			// err := DownloadVorteilFile("vtramp.img", "vtramp")
			if err != nil {
				return err
			}

			info, err = os.Stat(path)
			if err != nil {
				return err
			}

		} else {
			return err
		}

	}

	build.content.trampoline.first = build.content.kernel.last + 1
	build.content.trampoline.length = ceiling(uint64(info.Size()), SectorSize)
	build.content.trampoline.last = build.content.trampoline.first +
		build.content.trampoline.length - 1

	// app
	path = build.args.Binary

	info, err = os.Stat(path)
	if err != nil {
		return err
	}

	build.content.app.first = build.content.trampoline.last + 1
	build.content.app.length = ceiling(uint64(info.Size()), SectorSize)
	build.content.app.last = build.content.app.first +
		build.content.app.length - 1

	// files
	build.content.LBAs = uint64(build.config.Disk.DiskSize) * megabyte / SectorSize
	fsSectors := build.content.LBAs - build.content.app.last - build.content.reserved.length - 1
	build.files, err = CompileFilesystem(build.args.Files, fsSectors)
	if err != nil {
		return err
	}

	info, err = os.Stat(build.files.Name())
	if err != nil {
		return err
	}

	build.content.files.first = build.content.app.last + 1
	build.content.files.length = fsSectors
	build.content.files.last = build.content.files.first +
		build.content.files.length - 1

	// backup
	build.content.backup.first = build.content.files.last + 1
	build.content.backup.length = 34
	build.content.backup.last = build.content.backup.first +
		build.content.backup.length - 1

	return nil

}
