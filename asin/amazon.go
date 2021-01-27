package asin

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/proxy"
	"strings"
)

type Features struct {
	Url      string
	Title    string
	PhotoUrl string
}

func ExtractFeaturesByUrl(rawUrl string, proxies *proxy.Proxies) (features Features, err error) {
	features.Url = rawUrl

	var html string
	html, err = lookupByUrl(features.Url, proxies)
	if err != nil {
		return
	}

	reader := strings.NewReader(html)

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		return
	}

	features.Title = strings.TrimSpace(doc.Find("#title span").First().Text())
	features.Title = strings.Replace(features.Title, "£", "", -1)

	features.PhotoUrl, _ = doc.Find("#imgTagWrapperId img[data-old-hires]").First().Attr("data-old-hires")

	return
}
