package dao

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/call2mall/conn"
	"github.com/jackc/pgx/v4"
)

type Image struct {
	Bytes []byte
}

func (i Image) Hash() (hash string) {
	return fmt.Sprintf("%x", md5.Sum(i.Bytes))
}

func (i Image) Store() (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		return i.store(tx)
	})

	return
}

func (i Image) store(tx pgx.Tx) (err error) {
	if len(i.Bytes) == 0 {
		err = fmt.Errorf("image doesn't contain bytes")

		return
	}

	query := `insert into asin.image (hash, bytes) values ($1, $2) on conflict (hash) do nothing;`

	_, err = tx.Exec(context.Background(), query, i.Hash(), i.Bytes)

	return
}
