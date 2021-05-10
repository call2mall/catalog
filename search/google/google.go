package google

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/chrome"
	"github.com/chromedp/chromedp"
	"net/url"
	"strings"
	"time"
)

type Google struct {
}

func (s Google) Search(query string, b *chrome.Chrome) (urlList []string, err error) {
	queryUrl := url.URL{
		Scheme: "https",
		Host:   "www.google.com",
		Path:   "search",
		RawQuery: url.Values{
			"q": []string{query},
		}.Encode(),
	}

	var html string
	err = b.Run(queryUrl.String(), []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible("#search", chromedp.ByID),
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
	doc.Find("#search a[href]:not([class])").Each(func(i int, sel *goquery.Selection) {
		href = sel.AttrOr("href", "")

		if strings.HasPrefix(href, "http") {
			urlList = append(urlList, href)
		}
	})

	return
}
