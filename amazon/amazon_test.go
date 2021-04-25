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

func TestExtractOne(t *testing.T) {
	a := NewAmazon()

	var (
		meta Meta
		ok   bool
		err  error
	)
	meta, ok, err = a.Extract("B00009ZSXJ")
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("Extractor was not able to get meta data of the product")
	}

	fmt.Println(meta)
}

func TestExtractAll(t *testing.T) {
	asinList, err := dao.GetAllASIN()
	if err != nil {
		t.Error(err)
	}

	a := NewAmazon()

	var (
		meta Meta
		ok   bool
	)

	for _, asin := range asinList {
		meta, ok, err = a.Extract(asin)
		if err != nil {
			t.Fatal(err)
		}

		if !ok {
			continue
		}

		fmt.Println(meta)
	}
}
