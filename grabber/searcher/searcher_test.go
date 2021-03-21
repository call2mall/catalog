package searcher

import (
	"github.com/call2mall/conn"
	"github.com/leprosus/golang-config"
	"github.com/leprosus/golang-log"
	"testing"
)

func init() {
	log.Stdout(true)

	_ = config.Init("../../config.json")
	_ = conn.InitSQL(config.String("psql.user"), config.String("psql.pass"), config.String("psql.database"), config.String("psql.host"), config.UInt32("psql.port"))
}

func TestSearcher(t *testing.T) {
	threads := config.UInt32("threads.searcher")

	err := RunSearchASIN(uint(threads))
	if err != nil {
		t.Fatal(err)
	}
}
