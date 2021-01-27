package translator

import (
	"encoding/json"
	"fmt"
	"github.com/call2mall/catalog/proxy"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func Translate(text string, proxies *proxy.Proxies) (translation string, err error) {
	options := url.Values{}
	options.Set("text", text)
	options.Set("from", "de")
	options.Set("to", "ru")

	data := options.Encode()

	var req *http.Request
	req, err = http.NewRequest("POST", "https://www.bing.com/ttranslate", strings.NewReader(data))
	if err != nil {
		return
	}

	req.Header.Set("origin", "https://www.bing.com")
	req.Header.Set("accept-language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	req.Close = true

	client := http.Client{}

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

	var res *http.Response
	res, err = client.Do(req)
	if err != nil {
		return
	}

	defer func() {
		_ = res.Body.Close()
	}()

	var bs []byte
	bs, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("get unexpected code %d (%s)", res.StatusCode, res.Status)

		return
	}

	translation = string(bs)

	resStruct := struct {
		Translation string `json:"translationResponse"`
	}{}

	err = json.Unmarshal(bs, &resStruct)
	if err != nil {
		return
	}

	translation = resStruct.Translation

	return
}
