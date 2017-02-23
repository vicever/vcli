package converter

import (
	"io/ioutil"
	"os"

	"github.com/sisatech/sherlock"
	"github.com/sisatech/vcli/compiler"
)

func ExportDynamicVHD(in Convertible, path, kernel string, debug bool) (*os.File, error) {

	var vhd *os.File

	err := sherlock.Try(func() {

		var err error

		if path == "" {
			vhd, err = ioutil.TempFile("", "")
			sherlock.Check(err)
			sherlock.Check(vhd.Close())
			path = vhd.Name()
		}

		// create temp dir for files
		tmp, err := ioutil.TempDir("", "")
		sherlock.Check(err)

		sherlock.Check(ExportLoose(in, tmp))

		defer os.RemoveAll(tmp)

		// build sparse vmdk
		out, err := compiler.BuildDynamicVHD(tmp+"/app",
			tmp+"/app.vcfg", tmp+"/fs", kernel, path, debug)
		sherlock.Check(err)

		vhd, err = os.Open(out)
		sherlock.Check(err)

	})

	return vhd, err

}
