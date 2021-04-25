package swisscows

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/chrome"
	"github.com/chromedp/chromedp"
	"net/url"
	"strings"
	"time"
)

type SwissCows struct {
}

func (s SwissCows) Search(query string, b *chrome.Browser) (urlList []string, err error) {
	queryUrl := url.URL{
		Scheme: "https",
		Host:   "swisscows.com",
		Path:   "web",
		RawQuery: url.Values{
			"query":  []string{query},
			"region": []string{"en-US"},
		}.Encode(),
	}

	var html string
	err = b.Run(queryUrl.String(), []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible(".page-results"),
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
	doc.Find(".page-results a.title-wrap").Each(func(i int, sel *goquery.Selection) {
		href = sel.AttrOr("href", "")

		urlList = append(urlList, href)
	})

	return
}
