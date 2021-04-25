package lycos

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/chrome"
	"github.com/chromedp/chromedp"
	"net/url"
	"strings"
	"time"
)

type Lycos struct {
}

func (s Lycos) Search(query string, b *chrome.Browser) (urlList []string, err error) {
	queryUrl := url.URL{
		Scheme: "https",
		Host:   "search3.lycos.com",
		Path:   "web",
		RawQuery: url.Values{
			"q": []string{query},
		}.Encode(),
	}

	var html string
	err = b.Run(queryUrl.String(), []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible(".content.con-search"),
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
	doc.Find(".content.con-search a.result-link").Each(func(i int, sel *goquery.Selection) {
		href = sel.AttrOr("href", "")

		urlList = append(urlList, href)
	})

	return
}
