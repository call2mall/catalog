package dao

import (
	"crypto/md5"
	"fmt"
	"github.com/call2mall/storage/db"
	"github.com/jmoiron/sqlx"
)

type Image struct {
	Bytes []byte
}

func (i Image) Hash() (hash string) {
	return fmt.Sprintf("%x", md5.Sum(i.Bytes))
}

func (i Image) Store() (err error) {
	err = db.WithSQL(func(tx *sqlx.Tx) (err error) {
		return i.store(tx)
	})

	return
}

func (i Image) store(tx *sqlx.Tx) (err error) {
	if len(i.Bytes) == 0 {
		err = fmt.Errorf("image doesn't contain bytes")

		return
	}

	query := `insert into asin.image (hash, bytes) values ($1, $2) on conflict (hash) do nothing;`

	_, err = tx.Exec(query, i.Hash(), i.Bytes)

	return
}
