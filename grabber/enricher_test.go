package grabber

import (
	"fmt"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/proxy"
	"github.com/call2mall/conn"
	"github.com/leprosus/golang-config"
	"github.com/leprosus/golang-log"
	"testing"
)

func init() {
	_ = log.Path("../log")
	log.Stdout(true)

	_ = config.Init("../config.json")
	_ = conn.InitSQL(config.String("psql.user"), config.String("psql.pass"), config.String("psql.database"), config.String("psql.host"), config.UInt32("psql.port"))
}

func TestEnricher(t *testing.T) {
	err := RunEnrichASIN(10)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExtractASINMeta(t *testing.T) {
	var (
		number     = dao.ASIN("B07RLTM7ZL")
		originList dao.OriginList

		features dao.ASINFeatures
		ok       bool
		err      error
	)

	originList, err = number.LoadOrigins()
	if err != nil {
		t.Fatal(err)
	}

	var proxies *proxy.Proxies
	proxies, err = proxy.GetInstance()
	if err != nil {
		t.Fatal(err)
	}

	features, ok, err = extractASINMeta(number, originList, proxies)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatalf("Can't extract features from ASIN `%s`", number)
	}

	fmt.Println(features)
}
