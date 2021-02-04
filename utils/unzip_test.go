package utils

import (
	"io"
	"testing"
)

func TestUnzip(t *testing.T) {
	var count int
	err := Unzip("../docs/wetransfer-7446ad.zip", func(reader io.ReadCloser) (err error) {
		count++

		return
	})

	if err != nil {
		t.Fatal(err)
	}

	if count < 100 {
		t.Fatal("Can't unzip archive")
	}
}
