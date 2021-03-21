package webtran

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type WebTran struct {
	timestamp time.Duration
}

func (t *WebTran) Timestamp(timestamp time.Duration) {
	t.timestamp = timestamp
}
func (t WebTran) Translate(text, from, to, proxyAddr string) (result string, err error) {
	options := url.Values{}
	options.Set("text", text)
	options.Set("gfrom", from)
	options.Set("gto", to)

	var bs []byte
	bs, err = t.request("https://www.webtran.ru/gtranslate/", options, proxyAddr)
	if err != nil {
		return
	}

	result = string(bs)

	return
}

func (t WebTran) request(rawUrl string, options url.Values, proxyAddr string) (bs []byte, err error) {
	data := options.Encode()

	var req *http.Request
	req, err = http.NewRequest("POST", rawUrl, strings.NewReader(data))
	if err != nil {
		return
	}

	req.Header.Set("accept-language", "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7,de;q=0.6,lt;q=0.5,pl;q=0.4,zh-CN;q=0.3,zh;q=0.2")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.66 Safari/537.36")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("origin", "https://www.webtran.ru")
	req.Header.Set("referer", "https://www.webtran.ru/")
	req.Header.Set("x-requested-with", "XMLHttpRequest")

	req.Close = true

	client := http.Client{}

	if t.timestamp > 0 {
		client.Timeout = t.timestamp
	}

	if len(proxyAddr) > 0 {
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

	bs, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("get unexpected code %d (%s)", res.StatusCode, res.Status)

		return
	}

	return
}
