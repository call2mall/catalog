package bing

import (
	"fmt"
	"github.com/call2mall/catalog/chrome"
	"testing"
	"time"
)

func TestBingSearch(t *testing.T) {
	b := chrome.New()

	query := fmt.Sprintf("\"%s\"", "B07K3SS94V")

	var (
		list []string
		err  error
	)
	list, err = Bing{}.Search(query, b)
	if err != nil {
		t.Error(err)
	}

	if len(list) == 0 {
		t.Error("Can't extract links")
	}

	fmt.Println(list)

	time.Sleep(2 * time.Second)
}
