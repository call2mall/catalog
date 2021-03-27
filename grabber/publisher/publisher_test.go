package publisher

import (
	"github.com/call2mall/conn"
	"github.com/leprosus/golang-config"
	"github.com/leprosus/golang-log"
	"io/ioutil"
	"os"
	"testing"
)

func init() {
	log.Stdout(true)

	_ = config.Init("../../config.json")
	_ = conn.InitSQL(config.String("psql.user"), config.String("psql.pass"), config.String("psql.database"), config.String("psql.host"), config.UInt32("psql.port"))
}

func TestPublisher(t *testing.T) {
	threads := config.UInt32("threads.publisher")

	err := RunPublisher(uint(threads))
	if err != nil {
		t.Fatal(err)
	}
}

func TestPrepareImage(t *testing.T) {
	src, err := os.Open("../../docs/image.jpg")
	if err != nil {
		t.Error(err)
	}

	var in []byte
	in, err = ioutil.ReadAll(src)
	if err != nil {
		t.Error(err)
	}

	_, err = prepareImage(in, 200)
	if err != nil {
		t.Error(err)
	}
}
