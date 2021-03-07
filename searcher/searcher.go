package searcher

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/crome"
	"github.com/call2mall/catalog/proxy"
	"github.com/leprosus/golang-log"
	"net/url"
	"strings"
	"sync"
)

var domains = []string{
	"amazon.co.uk",
	"amazon.com.au",
	"amazon.co.jp",
	"amazon.sg",
	"amazon.nl",
	"amazon.se",
	"amazon.de",
	"amazon.fr",
	"amazon.es",
	"amazon.it",
	"amazon.ae",
}

func SearchByUrl(asin string, proxies *proxy.Proxies) (urlList []string, err error) {
	var (
		wg      = &sync.WaitGroup{}
		urlData url.URL

		resUrl string
		ok     bool

		pos int
	)

	for _, host := range domains {
		urlData = url.URL{
			Scheme: "https",
			Host:   host,
			Path:   fmt.Sprintf("/review/product/%s", asin),
		}

		wg.Add(1)

		go func(urlData url.URL) {
			defer wg.Done()

			resUrl, ok, err = searchByUrl(urlData, proxies)
			if err != nil {
				log.WarnFmt("Can't search review page for ASIN `%s`: %v", asin, err)
			}

			if ok {
				pos = strings.Index(resUrl, "/ref=")
				if pos > -1 {
					resUrl = resUrl[0:pos]
				}

				urlList = append(urlList, resUrl)
			}
		}(urlData)
	}

	wg.Wait()

	return
}

func searchByUrl(urlData url.URL, proxies *proxy.Proxies) (resUrl string, ok bool, err error) {
	var html string
	html, err = crome.GetHtml(urlData.String(), proxies)
	if err != nil {
		return
	}

	reader := strings.NewReader(html)

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		return
	}

	resUrl, ok = doc.Find(".product-info a.a-link-normal[href]").First().Attr("href")

	var resData *url.URL
	resData, err = url.Parse(resUrl)
	if err != nil {
		return
	}

	if !strings.HasPrefix(resUrl, "http") {
		resData = urlData.ResolveReference(resData)
	}

	resData.RawQuery = url.Values{}.Encode()

	resUrl = resData.String()

	return
}
