package dao

import (
	"fmt"
	"github.com/call2mall/conn"
	"github.com/jmoiron/sqlx"
	"math/rand"
	"net/url"
	"sort"
	"strings"
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
		country, _, _, err = DetectL8nOfOrigin(rawUrl)
		if err != nil {
			return
		}

		originList[country] = rawUrl
	}

	return
}

func DetectL8nOfOrigin(rawUrl string) (country, l8n string, withL8nSwitch bool, err error) {
	var urlData *url.URL
	urlData, err = url.Parse(rawUrl)
	if err != nil {
		return
	}

	urlData.Host = strings.Replace(urlData.Host, "www.", "", 1)
	urlData.Host = strings.Replace(urlData.Host, "amazon.", "", 1)

	switch urlData.Host {
	case "co.uk":
		country = "uk"
		l8n = "en"
	case "com.au":
		country = "au"
		l8n = "en"
	case "sg":
		country = "sg"
		l8n = "en"
	case "ae":
		country = "ae"
		l8n = "en"
		withL8nSwitch = true
	case "co.jp":
		country = "jp"
		l8n = "en"
		withL8nSwitch = true
	case "de":
		country = "de"
		l8n = "en"
		withL8nSwitch = true
	case "it":
		country = "it"
		l8n = "it"
	case "fr":
		country = "fr"
		l8n = "fr"
	case "es":
		country = "es"
		l8n = "es"
	case "nl":
		country = "nl"
		l8n = "en"
		withL8nSwitch = true
	case "se":
		country = "se"
		l8n = "sv"
	default:
		err = fmt.Errorf("can't detect localisation of `%s`", rawUrl)

		return
	}

	return
}

func CountryPriority() (list []string) {
	list = []string{"uk", "de", "au", "nl", "sg", "jp", "ae"}
	for i := 0; i < 10; i++ {
		sort.Slice(list, func(_, _ int) bool {
			return rand.Int31n(100) > 50
		})
	}

	list = append(list, "es", "fr", "it", "se")

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
