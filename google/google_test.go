package google

import (
	"fmt"
	"github.com/call2mall/catalog/proxy"
	"testing"
)

func TestFindPageByASIN(t *testing.T) {
	proxies := proxy.NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.10.102:29842",
		"http://emiles01:xVypbJnv@51.89.130.34:29842",
		"http://emiles01:xVypbJnv@51.83.17.111:29842",
		"http://emiles01:xVypbJnv@51.89.31.32:29842",
		"http://emiles01:xVypbJnv@51.89.131.103:29842",
	})

	urlList, err := FindPageByASIN("B001A1V4M6", proxies)
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Println(urlList)
}

func TestFindCachedPageByUrl(t *testing.T) {
	proxies := proxy.NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.10.102:29842",
		"http://emiles01:xVypbJnv@51.89.130.34:29842",
		"http://emiles01:xVypbJnv@51.83.17.111:29842",
		"http://emiles01:xVypbJnv@51.89.31.32:29842",
		"http://emiles01:xVypbJnv@51.89.131.103:29842",
	})

	pageUrl, err := FindPageByASIN("https://www.amazon.it/Werder-Bremen-SVW-Salvadanaio-Maialino/dp/B07K3SS94V", proxies)
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Println(pageUrl)
}
