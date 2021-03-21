package browser

import (
	"fmt"
	"testing"
)

func TestLookupByUrl(t *testing.T) {
	rawUrl := "https://www.amazon.de/review/product/B07Q3S8BKF"
	b := NewBrowser()
	err := b.Proxy("http://emiles01:xVypbJnv@51.89.130.34:29842")
	if err != nil {
		t.Error(err)
	}

	var html string
	html, err = b.GetHtml(rawUrl)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(html)
}
