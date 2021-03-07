package searcher

import (
	"fmt"
	"github.com/call2mall/catalog/proxy"
	"testing"
)

func TestSearchByUrl(t *testing.T) {
	proxies := proxy.NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.10.102:29842",
		"http://emiles01:xVypbJnv@51.89.130.34:29842",
		"http://emiles01:xVypbJnv@51.83.17.111:29842",
		"http://emiles01:xVypbJnv@51.89.31.32:29842",
		"http://emiles01:xVypbJnv@51.89.131.103:29842",
		"http://emiles01:xVypbJnv@51.77.236.120:29842",
		"http://emiles01:xVypbJnv@51.91.195.216:29842",
		"http://emiles01:xVypbJnv@51.91.196.233:29842",
		"http://emiles01:xVypbJnv@51.89.10.221:29842",
		"http://emiles01:xVypbJnv@51.89.130.55:29842",
	})

	urlList, err := SearchByUrl("B004K8K7MO", proxies)
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Println(urlList)
}
