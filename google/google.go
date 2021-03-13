package google

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/chrome"
	"github.com/call2mall/catalog/proxy"
	"net/url"
	"regexp"
	"strings"
)

var domainsRegExp = regexp.MustCompile("^(www\\.)?amazon\\.(co\\.uk|com\\.au|co\\.jp|ae|de|it|fr|es|nl|se|sg)$")

func FindPageByASIN(asin string, proxies *proxy.Proxies) (urlList []string, err error) {
	rawUrl := fmt.Sprintf("https://www.google.com/search?q=\"%s\"", asin)

	var html string
	html, err = chrome.GetHtml(rawUrl, proxies)
	if err != nil {
		return
	}

	reader := strings.NewReader(html)

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		return
	}

	var (
		href    string
		urlData *url.URL
		matches [][]string
	)
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href = sel.AttrOr("href", "")

		urlData, err = url.Parse(href)
		if err != nil {
			return
		}

		matches = domainsRegExp.FindAllStringSubmatch(urlData.Host, -1)
		if len(matches) == 0 {
			return
		}

		if !strings.Contains(urlData.Path, "/dp/") {
			return
		}

		urlList = append(urlList, href)
	})

	return
}
