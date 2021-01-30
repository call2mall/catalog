package asin

import (
	"fmt"
	"github.com/call2mall/catalog/proxy"
	"testing"
)

func TestExtractFeaturesByUrl(t *testing.T) {
	proxies := proxy.NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.10.102:29842",
		"http://emiles01:xVypbJnv@51.89.130.34:29842",
		"http://emiles01:xVypbJnv@51.83.17.111:29842",
		"http://emiles01:xVypbJnv@51.89.31.32:29842",
		"http://emiles01:xVypbJnv@51.89.131.103:29842",
	})

	features, err := ExtractFeaturesByUrl("https://www.amazon.co.uk/Brio-footmuff-Shiny-perambulator-pushchair/dp/B004LDFV1U", proxies)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(features.Url)
	fmt.Println(features.Title)
	fmt.Println(features.PhotoUrl)
	fmt.Println(features.Category)
}