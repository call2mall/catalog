package yahoo

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/browser"
	"github.com/chromedp/chromedp"
	"net/url"
	"strings"
	"time"
)

type Yahoo struct {
}

func (s Yahoo) Search(query string, b *browser.Browser) (urlList []string, err error) {
	queryUrl := url.URL{
		Scheme: "https",
		Host:   "search.yahoo.com",
		Path:   "search",
		RawQuery: url.Values{
			"p": []string{query},
		}.Encode(),
	}

	var html string
	err = b.Run(queryUrl.String(), []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible("#web", chromedp.ByID),
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
	doc.Find("#web a[target=_blank]").Each(func(i int, sel *goquery.Selection) {
		href = sel.AttrOr("href", "")

		urlList = append(urlList, href)
	})

	return
}
