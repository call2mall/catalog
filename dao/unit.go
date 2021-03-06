package dao

import (
	"context"
	"database/sql"
	"github.com/call2mall/conn"
	"github.com/jackc/pgx/v4"
)

type Unit struct {
	Id                uint64
	WarehouseId       string
	EAN               string
	ASIN              ASIN
	SKU               string
	Condition         Condition
	Quantity          uint32
	UnitCostInCent    uint32
	UnitDiscount      uint32
	RetailPriceInCent uint32
	IsPublished       bool
}

func (u *Unit) Store() (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		return u.store(tx)
	})

	return
}

func (u *Unit) store(tx pgx.Tx) (err error) {
	query := `insert into catalog.unit (warehouse_id, ean, asin, sku, condition, quantity, unit_cost, unit_discount, retail_price)
				values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
				on conflict (warehouse_id, ean, asin)
				do update set sku = $4, condition = $5, quantity = $6, unit_cost = $7, unit_discount = $7, retail_price = $8, is_remove = false
				returning id;`

	err = tx.QueryRow(context.Background(), query,
		u.WarehouseId,
		u.EAN,
		u.ASIN,
		sql.NullString{
			String: u.SKU,
			Valid:  len(u.SKU) > 0,
		},
		u.Condition,
		u.Quantity,
		u.UnitCostInCent,
		u.UnitDiscount,
		u.RetailPriceInCent,
	).Scan(&u.Id)
	if err != nil {
		return
	}

	return
}

type UnitList []Unit

func (ul UnitList) ExtractASINList() (l ASINList) {
	var (
		uniq = map[ASIN]interface{}{}
		ok   bool
	)
	for _, u := range ul {
		_, ok = uniq[u.ASIN]
		if ok {
			continue
		}

		uniq[u.ASIN] = nil

		l = append(l, u.ASIN)
	}

	return
}

func (ul UnitList) Store() (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		for _, unit := range ul {
			err = unit.store(tx)
			if err != nil {
				return
			}
		}

		return
	})

	return
}

func RemoveUnitListByASINList(al ASINList) (err error) {
	err = conn.WithSQL(func(tx pgx.Tx) (err error) {
		query := `delete from catalog.unit where asin = $1;`

		for _, a := range al {
			_, err = tx.Exec(context.Background(), query, a)
			if err != nil {
				return
			}
		}

		return
	})

	return
}
