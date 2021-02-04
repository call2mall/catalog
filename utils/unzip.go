package utils

import (
	"archive/zip"
	"io"
)

func Unzip(zipPath string, handler func(reader io.ReadCloser) (err error)) (err error) {
	var reader *zip.ReadCloser
	reader, err = zip.OpenReader(zipPath)
	if err != nil {
		return
	}
	defer func() {
		_ = reader.Close()
	}()

	var file io.ReadCloser
	for _, zipFile := range reader.File {
		if zipFile.FileInfo().IsDir() {
			continue
		}

		file, err = zipFile.Open()
		if err != nil {
			return
		}

		err = handler(file)
		if err != nil {
			_ = file.Close()

			return
		}

		_ = file.Close()
	}

	return
}
