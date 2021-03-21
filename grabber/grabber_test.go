package grabber

import (
	"github.com/call2mall/conn"
	"github.com/leprosus/golang-config"
	"github.com/leprosus/golang-log"
	"testing"
)

func init() {
	log.Stdout(true)

	_ = config.Init("../config.json")
	_ = conn.InitSQL(config.String("psql.user"), config.String("psql.pass"), config.String("psql.database"), config.String("psql.host"), config.UInt32("psql.port"))
}

func TestGrab(t *testing.T) {
	err := GrabUnits()
	if err != nil {
		t.Fatal(err)
	}
}
