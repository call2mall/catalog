package amazon

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

	features, ok, err := ExtractFeaturesByUrl("https://www.amazon.co.uk/Garmin-Instinct-Features-Monitoring-Lakeside/dp/B07PN8C9V2", proxies)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("Can't extract features")
	}

	fmt.Println(features.Url)
	fmt.Println(features.Title)
	fmt.Println(features.PhotoUrl)
	fmt.Println(features.Category)
}

func TestExtractFeaturesByCache(t *testing.T) {
	proxies := proxy.NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.10.102:29842",
		"http://emiles01:xVypbJnv@51.89.130.34:29842",
		"http://emiles01:xVypbJnv@51.83.17.111:29842",
		"http://emiles01:xVypbJnv@51.89.31.32:29842",
		"http://emiles01:xVypbJnv@51.89.131.103:29842",
	})

	features, ok, err := ExtractFeaturesByUrl("https://www.amazon.com/Silkn-Flash-Go-Express-Permanent/dp/B01AR6U6D2", proxies)
	if err != nil {
		t.Fatal(err)
	}

	if !ok {
		t.Fatal("Can't extract features")
	}

	fmt.Println(features.Url)
	fmt.Println(features.Title)
	fmt.Println(features.PhotoUrl)
	fmt.Println(features.Category)
}
