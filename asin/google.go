package asin

import (
	"fmt"
	"github.com/call2mall/catalog/proxy"
	"regexp"
	"strings"
)

var domainsRegExp = regexp.MustCompile("amazon(\\.co)?\\.(uk|de|it|fr|es)")

func FindPageByASIN(asin string, proxies *proxy.Proxies) (urlMap map[string]string, err error) {
	urlMap = map[string]string{}

	var html string
	html, err = lookupByGoogle(asin, proxies)
	if err != nil {
		return
	}

	pattern := fmt.Sprintf("\\bhttp(s)?://(www\\.)?amazon\\.[a-z\\.]{1,5}/[^/]+/dp/%s\\b", asin)
	regExp := regexp.MustCompile(pattern)

	list := regExp.FindAllString(string(html), -1)

	unique := map[string]struct{}{}

	for _, u := range list {
		if !strings.Contains(u, "...") {
			unique[u] = struct{}{}
		}
	}

	var (
		data   []string
		domain string
	)
	for u := range unique {
		data = domainsRegExp.FindStringSubmatch(u)

		if data != nil {
			domain = data[2]

			urlMap[domain] = u
		}
	}

	return
}

func lookupByGoogle(asin string, proxies *proxy.Proxies) (html string, err error) {
	rawUrl := fmt.Sprintf("https://www.google.com/search?q=%s", asin)

	return lookupByUrl(rawUrl, proxies)
}
