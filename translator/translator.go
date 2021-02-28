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

type translateRes struct {
	Result         string `json:"result"`
	TranslatedText string `json:"translated_text"`
}

func Translate(text, from, to string, proxies *proxy.Proxies) (translation string, err error) {
	options := url.Values{}
	options.Set("text_to_translate", text)
	options.Set("source_lang", from)
	options.Set("translated_lang", to)
	options.Set("use_cache_only", "false")

	var bs []byte
	bs, err = request("https://www.translate.com/translator/ajax_translate", options, proxies)
	if err != nil {
		return
	}

	var answer translateRes
	err = json.Unmarshal(bs, &answer)
	if err != nil {
		return
	}

	if answer.Result != "success" {
		err = fmt.Errorf("can't translate `%s` because get the answer: %s", text, string(bs))

		return
	}

	translation = answer.TranslatedText

	return
}

type detectionRes struct {
	Result   string `json:"result"`
	Language string `json:"language"`
}

func DetectLang(text string, proxies *proxy.Proxies) (lang string, err error) {
	options := url.Values{}
	options.Set("text_to_translate", text)

	var bs []byte
	bs, err = request("https://www.translate.com/translator/ajax_lang_auto_detect", options, proxies)
	if err != nil {
		return
	}

	var answer detectionRes
	err = json.Unmarshal(bs, &answer)
	if err != nil {
		return
	}

	if answer.Result != "success" {
		err = fmt.Errorf("can't detect language of `%s` because get the answer: %s", text, string(bs))

		return
	}

	lang = answer.Language

	return
}

func request(rawUrl string, options url.Values, proxies *proxy.Proxies) (bs []byte, err error) {
	data := options.Encode()

	var req *http.Request
	req, err = http.NewRequest("POST", rawUrl, strings.NewReader(data))
	if err != nil {
		return
	}

	req.Header.Set("origin", "https://www.bing.com")
	req.Header.Set("accept-language", "en-US,en;q=0.9,ru-RU;q=0.8,ru;q=0.7,de;q=0.6,lt;q=0.5,pl;q=0.4,zh-CN;q=0.3,zh;q=0.2")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.66 Safari/537.36")
	req.Header.Set("content-type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("origin", "https://www.translate.com")
	req.Header.Set("referer", "https://www.translate.com/")
	req.Header.Set("x-requested-with", "XMLHttpRequest")

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
