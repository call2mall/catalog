package dowloader

import (
	"github.com/call2mall/catalog/proxy"
	"testing"
)

func TestDownloadFromWetransfer(t *testing.T) {
	proxies := proxy.NewProxies([]string{
		"http://emiles01:xVypbJnv@51.89.10.102:29842",
		"http://emiles01:xVypbJnv@51.89.130.34:29842",
		"http://emiles01:xVypbJnv@51.83.17.111:29842",
		"http://emiles01:xVypbJnv@51.89.31.32:29842",
		"http://emiles01:xVypbJnv@51.89.131.103:29842",
	})

	err := DownloadFromWetransfer("https://we.tl/t-97j6I5ZDF2", proxies, ".")
	if err != nil {
		t.Fatal(err.Error())
	}
}
