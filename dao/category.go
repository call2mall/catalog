package dao

import (
	"context"
	"fmt"
	"github.com/call2mall/conn"
	"github.com/jackc/pgx/v4"
)

type Category struct {
	Id   uint32
	Name []string
}

func (c *Category) Store() (id uint32, err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		id, err = c.store(tx)

		return
	})

	return
}

func (c *Category) store(tx pgx.Tx) (id uint32, err error) {
	if len(c.Name) == 0 {
		err = fmt.Errorf("category name is empty")

		return
	}

	if c.Id > 0 {
		upd := `update asin.category set name = $2 where id = $1;`

		_, err = tx.Exec(context.Background(), upd, c.Id, c.Name)
		if err != nil {
			return
		}
	} else {
		sel := `select c.id from asin.category c where c.name = $1;`

		err = tx.QueryRow(context.Background(), sel, c.Name).Scan(&id)
		if err == pgx.ErrNoRows {
			ins := `insert into asin.category (name) values ($1) returning id;`

			err = tx.QueryRow(context.Background(), ins, c.Name).Scan(&id)
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
