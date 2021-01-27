package curl

import (
	"fmt"
	"github.com/call2mall/catalog/proxy"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func LoadByUrl(rawUrl string, header http.Header, proxies *proxy.Proxies) (bs []byte, state uint, err error) {
	var (
		req *http.Request
		res *http.Response
	)

	client := http.Client{
		Timeout: time.Minute,
	}

	proxyAddr, ok := proxies.Next()
	if ok {
		var proxyUrl *url.URL

		proxyUrl, err = url.Parse(proxyAddr)
		if err != nil {
			return
		}

		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	req, err = http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		return
	}

	req.Header = header
	req.Close = true

	res, err = client.Do(req)
	if err != nil {
		return
	}

	state = uint(res.StatusCode)

	switch res.StatusCode {
	case 404:
		bs = []byte{}
	case 200:
		bs, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return
		}

		err = res.Body.Close()
		if err != nil {
			err = nil
		}
	default:
		err = fmt.Errorf("get unexpected status code %d (%s)", res.StatusCode, res.Status)

		return
	}

	return
}
