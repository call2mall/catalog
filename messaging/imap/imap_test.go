package main

import (
	"github.com/leprosus/golang-config"
	"testing"
)

func init() {
	_ = config.Init("../config.json")
}

func TestImap(t *testing.T) {
	imap, err := NewClient(config.String("imap.username"), config.String("imap.password"), config.String("imap.host"), uint(config.Int32("imap.port")))
	if err != nil {
		t.Fatal(err)
	}

	imap.SetSearchMark(config.String("imap.mark"))

	defer func() {
		_ = imap.Close()
	}()

	_, err = imap.GetLatestEmail(0)
	if err != nil {
		t.Fatal(err)
	}
}
