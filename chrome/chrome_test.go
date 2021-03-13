package chrome

import (
	"fmt"
	"github.com/call2mall/catalog/proxy"
	"testing"
)

func TestLookupByUrl(t *testing.T) {
	proxies := proxy.NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.10.102:29842",
		"http://emiles01:xVypbJnv@51.89.130.34:29842",
		"http://emiles01:xVypbJnv@51.83.17.111:29842",
		"http://emiles01:xVypbJnv@51.89.31.32:29842",
		"http://emiles01:xVypbJnv@51.89.131.103:29842",
	})

	rawUrl := "https://www.amazon.de/review/product/B07Q3S8BKF"
	html, err := GetHtml(rawUrl, proxies)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(html)
}
