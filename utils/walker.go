package utils

import (
	"os"
	. "path"
	"path/filepath"
)

func Walk(path, extension string) (list []string, err error) {
	extension = "." + extension

	err = filepath.Walk(path, func(path string, info os.FileInfo, e error) (err error) {
		if e != nil {
			return e
		}

		if info.IsDir() {
			return
		}

		if filepath.Ext(path) != extension {
			return
		}

		if info.Size() == 0 {
			return
		}

		path, err = filepath.Abs(Clean(path) + "/")
		if err != nil {
			return
		}

		list = append(list, path)

		return nil
	})

	return
}
