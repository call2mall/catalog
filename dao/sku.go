package dao

import (
	"database/sql"
	"github.com/call2mall/catalog/category"
	"github.com/jmoiron/sqlx"
)

type SKU struct {
	Id                 int64
	EAN                string
	WarehouseId        string
	AvidesSKU          string
	Category           category.Category
	Title              string
	ASIN               string
	L8n                string
	Condition          category.Condition
	Quantity           uint64
	UnitCostInCent     uint64
	UnitDiscountInCent uint64
	RetailPriceInCent  uint64
	Image              Image
}

type SKUList []SKU

func (list SKUList) Store() (err error) {
	imageIns := `insert into image (hash, image) values ($1, $2) on conflict (hash) do nothing;`

	categorySel := `select c.id from category c where c.name = $1;`
	categoryIns := `insert into category (name, is_defined) values ($1, $2) returning id;`

	skuIns := `insert into sku (ean, warehouse_id, avides_sku, category_id, title, asin, condition, l8n, quantity, unit_cost, unit_discount, retail_price, image_hash)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
on conflict (ean, warehouse_id, avides_sku)
do update set category_id = $4, title = $5, asin = $6, condition = $7, l8n = $8, quantity = $9, unit_cost = $10, unit_discount = $11, retail_price = $12, image_hash = $13, timestamp = current_timestamp;`

	err = WithTx(func(tx *sqlx.Tx) (err error) {
		var hash sql.NullString

		for _, sku := range list {
			err = tx.QueryRowx(categorySel, sku.Category.Name).Scan(&sku.Category.Id)
			if err == sql.ErrNoRows {
				err = tx.QueryRow(categoryIns,
					sku.Category.Name,
					sku.Category.IsDefined,
				).Scan(&sku.Category.Id)
				if err != nil {
					return
				}
			} else if err != nil {
				return
			}

			hash = sql.NullString{}
			if len(sku.Image) > 0 {
				hash = sql.NullString{
					Valid:  true,
					String: sku.Image.Hash(),
				}

				_, err = tx.Exec(imageIns,
					sku.Image.Hash(),
					sku.Image,
				)
				if err != nil {
					return
				}
			}

			_, err = tx.Exec(skuIns,
				sku.EAN,
				sku.WarehouseId,
				sku.AvidesSKU,
				sku.Category.Id,
				sku.Title,
				sku.ASIN,
				sku.Condition,
				sku.L8n,
				sku.Quantity,
				sku.UnitCostInCent,
				sku.UnitDiscountInCent,
				sku.RetailPriceInCent,
				hash,
			)
			if err != nil {
				return
			}
		}

		return
	})

	return
}