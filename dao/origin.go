package dao

import (
	"github.com/call2mall/catalog/translate"
	"github.com/call2mall/conn"
	"github.com/jmoiron/sqlx"
	"math/rand"
	"sort"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type OriginList map[string]string

func UrlsToOriginList(list []string) (originList OriginList, err error) {
	originList = OriginList{}

	var country string
	for _, rawUrl := range list {
		country, _, _, err = translate.DetectL8nByAmazonUrl(rawUrl)
		if err != nil {
			return
		}

		originList[country] = rawUrl
	}

	return
}

func CountryPriority() (list []string) {
	list = []string{"ca", "uk", "de", "au", "nl", "sg", "jp", "ae", "sa", "in"}
	for i := 0; i < 10; i++ {
		sort.Slice(list, func(_, _ int) bool {
			return rand.Int31n(100) > 50
		})
	}

	list = append(list, "es", "fr", "it", "se", "mx")

	return
}

func (o OriginList) Store(asin ASIN) (err error) {
	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `insert into asin.origin (asin, country, url) values ($1, $2, $3) on conflict (asin, country) do nothing;`

		for country, rawUrl := range o {
			_, err = tx.Exec(query, asin, country, rawUrl)
			if err != nil {
				return
			}
		}

		return
	})

	return
}

func (a ASIN) LoadOrigins() (list OriginList, err error) {
	list = OriginList{}

	err = conn.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `select country, url from asin.origin where asin = $1;`

		var rows *sqlx.Rows
		rows, err = tx.Queryx(query, a)
		if err != nil {
			return
		}

		defer func() {
			_ = rows.Close()
		}()

		var country, rawUrl string
		for rows.Next() {
			err = rows.Scan(&country, &rawUrl)
			if err != nil {
				return
			}

			list[country] = rawUrl
		}

		return
	})

	return
}
