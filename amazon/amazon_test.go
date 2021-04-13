package amazon

import (
	"fmt"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/conn"
	config "github.com/leprosus/golang-config"
	log "github.com/leprosus/golang-log"
	"testing"
)

func init() {
	log.Stdout(true)

	_ = config.Init("../config.json")
	_ = conn.InitSQL(config.String("psql.user"), config.String("psql.pass"), config.String("psql.database"), config.String("psql.host"), config.UInt32("psql.port"))

	log.Path(config.Path("log.path"))
}

func TestParser(t *testing.T) {
	asinList, err := dao.GetAllASIN()
	if err != nil {
		t.Error(err)
	}

	var (
		asin dao.ASIN
		ch   = make(chan dao.ASIN)
	)

	for i := 0; i < 5; i++ {
		go instance(ch)
	}

	for _, asin = range asinList {
		ch <- asin
	}
}

func instance(ch chan dao.ASIN) {
	var (
		asin dao.ASIN

		meta Meta
		ok   bool
		err  error

		a *Amazon
	)
	for asin = range ch {
		a, err = NewAmazon()
		if err != nil {
			log.ErrorFmt("Can't init amazon handler: %v", err)

			continue
		}

		meta, ok, err = a.ExtractFromReview(asin)
		if err != nil {
			log.Error(err.Error())

			continue
		}

		if !ok {
			meta, ok, err = a.ExtractFromQA(asin)
			if err != nil {
				log.Error(err.Error())

				continue
			}
		}

		if !ok {
			meta, ok, err = a.ExtractFromProduct(asin)
			if err != nil {
				log.Error(err.Error())

				continue
			}
		}

		if ok {
			fmt.Println(meta.Title)
		}
	}
}
