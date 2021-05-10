package duckduckgo

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/chrome"
	"github.com/chromedp/chromedp"
	"net/url"
	"strings"
	"time"
)

type DuckDuckGo struct {
}

func (s DuckDuckGo) Search(query string, b *chrome.Chrome) (urlList []string, err error) {
	queryUrl := url.URL{
		Scheme: "https",
		Host:   "duckduckgo.com",
		RawQuery: url.Values{
			"q": []string{query},
		}.Encode(),
	}

	var html string
	err = b.Run(queryUrl.String(), []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible("#links", chromedp.ByID),
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
	doc.Find("#links a.result__a").Each(func(i int, sel *goquery.Selection) {
		href = sel.AttrOr("href", "")

		urlList = append(urlList, href)
	})

	return
}
