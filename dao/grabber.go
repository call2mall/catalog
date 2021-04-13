package dao

import (
	"context"
	"database/sql"
	"github.com/call2mall/conn"
	"github.com/jackc/pgx/v4"
)

func InsertWetransferUrl(uid uint32, rawUrl string) (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		query := `insert into asin.grabber (uid, url) values ($1, $2) on conflict (url) do nothing;`

		_, err = tx.Exec(context.Background(), query, uid, rawUrl)

		return
	})

	return
}

func GetLastEmailUID() (uid uint32, err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		query := `select max(uid) from asin.grabber;`

		var nullUID sql.NullInt64
		err = tx.QueryRow(context.Background(), query).Scan(&nullUID)

		if nullUID.Valid {
			uid = uint32(nullUID.Int64)
		}

		return
	})

	return
}

func IsProcessedWetransferUrl(rawUrl string) (ok bool, err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		query := `select count(g.uid) from asin.grabber g where g.url = $1;`

		var count uint
		err = tx.QueryRow(context.Background(), query, rawUrl).Scan(&count)
		if err != nil {
			return
		}

		ok = count > 0

		return
	})

	return
}
