package yahoo

import (
	"fmt"
	"github.com/call2mall/catalog/browser"
	"testing"
)

func TestSwissCowsSearch(t *testing.T) {
	b := browser.NewBrowser()
	err := b.Proxy("http://emiles01:xVypbJnv@51.89.130.34:29842")
	if err != nil {
		t.Error(err)
	}

	query := fmt.Sprintf("\"%s\"", "B07K3SS94V")

	var list []string
	list, err = Yahoo{}.Search(query, b)
	if err != nil {
		t.Error(err)
	}

	if len(list) == 0 {
		t.Error("Can't extract links")
	}

	fmt.Println(list)
}
