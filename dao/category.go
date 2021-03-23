package dao

import (
	"database/sql"
	"fmt"
	"github.com/call2mall/conn"
	"github.com/jmoiron/sqlx"
)

type Category struct {
	Id   uint32
	Name string
}

func (c *Category) Store() (id uint32, err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		id, err = c.store(tx)

		return
	})

	return
}

func (c *Category) store(tx *sqlx.Tx) (id uint32, err error) {
	if len(c.Name) == 0 {
		err = fmt.Errorf("category name is empty")

		return
	}

	if c.Id > 0 {
		upd := `update asin.category set name = $2 where id = $1;`

		_, err = tx.Exec(upd, c.Id, c.Name)
		if err != nil {
			return
		}
	} else {
		sel := `select c.id from asin.category c where c.name = $1;`

		err = tx.QueryRowx(sel, c.Name).Scan(&id)
		if err == sql.ErrNoRows {
			ins := `insert into asin.category (name) values ($1) returning id;`

			err = tx.QueryRow(ins, c.Name).Scan(&id)
			if err != nil {
				return
			}
		} else if err != nil {
			return
		}
	}

	c.Id = id

	return
}
