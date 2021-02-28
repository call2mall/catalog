package dao

import (
	"database/sql"
	"github.com/call2mall/storage/db"
	"github.com/jmoiron/sqlx"
)

func InsertWetransferUrl(uid uint32, rawUrl string) (err error) {
	err = db.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `insert into catalog.grabber (uid, url) values ($1, $2) on conflict (url) do nothing;`

		_, err = tx.Exec(query, uid, rawUrl)

		return
	})

	return
}

func GetLastEmailUID() (uid uint32, err error) {
	err = db.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `select max(uid) from catalog.grabber;`

		var nullUID sql.NullInt64
		err = tx.QueryRowx(query).Scan(&nullUID)

		if nullUID.Valid {
			uid = uint32(nullUID.Int64)
		}

		return
	})

	return
}

func IsProcessedWetransferUrl(rawUrl string) (ok bool, err error) {
	err = db.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `select count(g.uid) from catalog.grabber g where g.url = $1;`

		var count uint
		err = tx.QueryRowx(query, rawUrl).Scan(&count)
		if err != nil {
			return
		}

		ok = count > 0

		return
	})

	return
}
