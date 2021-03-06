package bing

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/chrome"
	"github.com/chromedp/chromedp"
	"net/url"
	"strings"
	"time"
)

type Bing struct {
}

func (s Bing) Search(query string, b *chrome.Chrome) (urlList []string, err error) {
	queryUrl := url.URL{
		Scheme: "https",
		Host:   "www.bing.com",
		Path:   "search",
		RawQuery: url.Values{
			"q": []string{query},
		}.Encode(),
	}

	var html string
	err = b.Run(queryUrl.String(), []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible("#b_results", chromedp.ByID),
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
	doc.Find("#b_results h2 a").Each(func(i int, sel *goquery.Selection) {
		href = sel.AttrOr("href", "")

		urlList = append(urlList, href)
	})

	return
}
