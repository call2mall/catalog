package amazon

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/crome"
	"github.com/call2mall/catalog/proxy"
	"net/url"
	"strings"
)

type Features struct {
	Url      string
	Title    string
	PhotoUrl string
	Category string
}

func ExtractFeaturesByUrl(rawUrl string, proxies *proxy.Proxies) (features Features, ok bool, err error) {
	cacheUrl := url.URL{
		Scheme: "https",
		Host:   "webcache.googleusercontent.com",
		Path:   "search",
		RawQuery: url.Values{
			"q": []string{fmt.Sprintf("cache:%s", rawUrl)},
		}.Encode(),
	}

	urlList := []string{rawUrl, cacheUrl.String()}

	for _, curUrl := range urlList {
		features, ok, err = extractFeaturesByUrl(curUrl, proxies)
		if !ok {
			continue
		}

		if err != nil {
			continue
		}
	}

	return
}

func extractFeaturesByUrl(rawUrl string, proxies *proxy.Proxies) (features Features, ok bool, err error) {
	features.Url = rawUrl

	var html string
	html, err = crome.GetHtml(features.Url, proxies)
	if err != nil {
		return
	}

	reader := strings.NewReader(html)

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		return
	}

	features.Title = doc.Find("#title span").First().Text()
	features.Title = strings.TrimSpace(features.Title)
	features.Title = strings.Replace(features.Title, " ", "", -1)
	features.Title = strings.Replace(features.Title, "£", "", -1)

	features.PhotoUrl, _ = doc.Find("#imgTagWrapperId img[src]").First().Attr("src")

	features.Category = doc.Find("#wayfinding-breadcrumbs_feature_div ul li:first-of-type a").First().Text()
	features.Category = strings.TrimSpace(features.Category)

	ok = len(features.Title) > 0 && len(features.PhotoUrl) > 0 && len(features.Category) > 0

	return
}
