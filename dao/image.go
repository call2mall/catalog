package dao

import (
	"crypto/md5"
	"fmt"
)

type Image []byte

func (im Image) Hash() (hash string) {
	return fmt.Sprintf("%x", md5.Sum(im))
}
