package dao

import (
	"fmt"
	"github.com/call2mall/storage/db"
	"github.com/jmoiron/sqlx"
	"net/url"
	"strings"
)

type OriginList map[string]string

func ListToOriginList(list []string) (originList OriginList, err error) {
	originList = OriginList{}

	var (
		urlData *url.URL
		lang    string
	)
	for _, rawUrl := range list {
		urlData, err = url.Parse(rawUrl)
		if err != nil {
			return
		}

		urlData.Host = strings.Replace(urlData.Host, "www.", "", 1)
		urlData.Host = strings.Replace(urlData.Host, "amazon.", "", 1)

		switch urlData.Host {
		case "co.uk", "co.au", "sg":
			lang = "en"
		case "ae":
			lang = "ar"
		case "co.jp":
			lang = "jp"
		case "de":
			lang = "de"
		case "it":
			lang = "it"
		case "fr":
			lang = "fr"
		case "es":
			lang = "es"
		case "nl":
			lang = "nl"
		case "se":
			lang = "sv"
		default:
			err = fmt.Errorf("can't detect language of `%s`", rawUrl)

			return
		}

		originList[lang] = rawUrl
	}

	return
}

func (o OriginList) Store(asin ASIN) (err error) {
	err = db.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `insert into asin.origin (asin, lang, url) values ($1, $2, $3) on conflict (asin, lang) do nothing;`

		for lang, rawUrl := range o {
			_, err = tx.Exec(query, asin, lang, rawUrl)
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

	err = db.WithSQL(func(tx *sqlx.Tx) (err error) {
		query := `select lang, url from asin.origin where asin = $1;`

		var rows *sqlx.Rows
		rows, err = tx.Queryx(query, a)
		if err != nil {
			return
		}

		defer func() {
			_ = rows.Close()
		}()

		var lang, rawUrl string
		for rows.Next() {
			err = rows.Scan(&lang, &rawUrl)
			if err != nil {
				return
			}

			list[lang] = rawUrl
		}

		return
	})

	return
}
