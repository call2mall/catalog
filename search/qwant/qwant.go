package qwant

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/chrome"
	"github.com/chromedp/chromedp"
	"net/url"
	"strings"
	"time"
)

type Qwant struct {
}

func (s Qwant) Search(query string, b *chrome.Browser) (urlList []string, err error) {
	queryUrl := url.URL{
		Scheme: "https",
		Host:   "www.qwant.com",
		RawQuery: url.Values{
			"q": []string{query},
		}.Encode(),
	}

	var html string
	err = b.Run(queryUrl.String(), []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible(".result"),
		chromedp.OuterHTML("html", &html),
	})
	if err != nil {
		return
	}

	reader := strings.NewReader(html)

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		return
	}

	var href string
	doc.Find(".results-column a.result--web--link").Each(func(i int, sel *goquery.Selection) {
		href = sel.AttrOr("href", "")

		urlList = append(urlList, href)
	})

	return
}
