package amazon

import (
	"fmt"
	"github.com/call2mall/catalog/browser"
	"github.com/call2mall/catalog/proxy"
	"testing"
)

func TestAmazon_FindPages(t *testing.T) {
	proxies := proxy.NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.10.102:29842",
		"http://emiles01:xVypbJnv@51.89.130.34:29842",
		"http://emiles01:xVypbJnv@51.83.17.111:29842",
		"http://emiles01:xVypbJnv@51.89.31.32:29842",
		"http://emiles01:xVypbJnv@51.89.131.103:29842",
	})

	a := Amazon{}
	urlList, err := a.FindPages("B07K3SS94V", proxies)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(urlList)
}

func TestAmazon_ExtractProps(t *testing.T) {
	proxies := proxy.NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.10.102:29842",
		"http://emiles01:xVypbJnv@51.89.130.34:29842",
		"http://emiles01:xVypbJnv@51.83.17.111:29842",
		"http://emiles01:xVypbJnv@51.89.31.32:29842",
		"http://emiles01:xVypbJnv@51.89.131.103:29842",
	})

	a := Amazon{}

	props, err := a.ExtractProps("https://www.amazon.it/PLAY-Bang-Olufsen-Beoplay-H4/dp/B07B6NRC7X", proxies)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(props)
}

func TestSearchThroughProductReport(t *testing.T) {
	b := browser.NewBrowser()

	urlList, err := searchThroughProductReport("B004K8K7MO", b)
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Println(urlList)
}
