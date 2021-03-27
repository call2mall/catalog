package dao

import (
	"github.com/call2mall/conn"
	"github.com/jmoiron/sqlx"
)

func (a ASIN) Publish() (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `update catalog.unit set is_published = true where asin = ?;`
		_, err = tx.Exec(query, a)

		return
	})

	return
}
