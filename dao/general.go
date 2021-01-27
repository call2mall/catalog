package dao

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/leprosus/golang-config"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"gitlab.evalab.ru/evalab/pornsci/entity"
	"sync"
)

var (
	instance *sqlx.DB
	once     = &sync.Once{}
)

func getPSQL() (db *sqlx.DB, err error) {
	once.Do(func() {
		dsn := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable", config.String("psql.user"), config.String("psql.pass"), config.String("psql.host"), config.Int64("psql.port"), config.String("psql.database"))

		instance, err = sqlx.Open("postgres", dsn)
		if err != nil {
			return
		}

		err = instance.Ping()
		if err != nil {
			return
		}
	})

	return instance, nil
}

func Close() (err error) {
	var db *sqlx.DB
	db, err = getPSQL()
	if err != nil {
		return
	}

	return db.Close()
}

func WithTx(fn func(tx *sqlx.Tx) (err error)) (err error) {
	var db *sqlx.DB
	db, err = getPSQL()
	if err != nil {
		return
	}

	var tx *sqlx.Tx
	tx, err = db.Beginx()
	if err != nil {
		return
	}

	defer func() {
		var e error
		if err != nil {
			e = tx.Rollback()
			if e != nil {
				err = errors.Wrap(err, e.Error())
			}
		} else {
			e = tx.Commit()
			if e != nil {
				err = errors.Wrap(err, e.Error())
			}
		}
	}()

	err = fn(tx)

	return
}

func SetSchema(tx *sqlx.Tx, schema entity.Schema) (err error) {
	setQuery := `set search_path = "%s";`

	_, err = tx.Exec(fmt.Sprintf(setQuery, schema))

	return
}
