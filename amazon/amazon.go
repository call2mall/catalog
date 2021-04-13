package amazon

import (
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	. "github.com/call2mall/catalog/browser"
	"github.com/call2mall/catalog/curl"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/proxy"
	"github.com/call2mall/catalog/user_agent"
	"github.com/chromedp/chromedp"
	log "github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Amazon struct {
	domains []string

	browser   *Browser
	proxies   *proxy.Proxies
	userAgent *user_agent.Rotator
}

type Meta struct {
	Url      string
	Title    string
	Category []string
	Bytes    []byte
}

var (
	NoProductCategory  = errors.New("there isn't product category")
	NoProductTitle     = errors.New("there isn't product title")
	NoProductImage     = errors.New("there isn't product image")
	DetectedAutomation = errors.New("detected automation")
)

func NewAmazon() (a *Amazon, err error) {
	browser := NewBrowser()
	browser.AddAcceptedResponseCode(503)

	var proxies *proxy.Proxies
	proxies, err = proxy.GetInstance()
	if err != nil {
		return
	}

	a = &Amazon{
		domains: []string{
			"amazon.com",
			"amazon.ca",
			"amazon.co.uk",
			"amazon.com.au",
			"amazon.sg",
		},

		browser:   browser,
		proxies:   proxies,
		userAgent: user_agent.GetInstance(),
	}

	return
}

func (a *Amazon) ExtractMeta(html string) (meta Meta, err error) {
	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return
	}

	meta.Title = doc.Find("#title span").First().Text()
	meta.Title = strings.TrimSpace(meta.Title)
	meta.Title = strings.Replace(meta.Title, " ", "", -1)
	if len(meta.Title) == 0 {
		err = NoProductTitle

		return
	}

	doc.Find("#wayfinding-breadcrumbs_feature_div ul li a").Each(func(i int, sel *goquery.Selection) {
		var category = sel.Text()
		category = strings.TrimSpace(category)
		category = strings.ReplaceAll(category, "&amp;", "and")

		category = strings.ToUpper(category[0:1]) + category[1:]

		meta.Category = append(meta.Category, category)
	})

	if len(meta.Category) == 0 {
		err = NoProductCategory

		return
	}

	photoUrl, ok := doc.Find("#imgTagWrapperId img[src]").First().Attr("src")
	if !ok {
		err = NoProductImage

		return
	}

	header := http.Header{}
	header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")

	var proxies *proxy.Proxies
	proxies, err = proxy.GetInstance()
	if err != nil {
		return
	}

	var state uint
	meta.Bytes, state, err = curl.LoadByUrl(photoUrl, header, proxies)
	if err != nil {
		return
	} else if state != 200 {
		err = fmt.Errorf("can't load image from `%s` because the server returned unexpected state", photoUrl)

		return
	} else if len(meta.Bytes) == 0 {
		err = fmt.Errorf("it returns zero-side image from `%s`", photoUrl)

		return
	}

	return
}

func (a *Amazon) ExtractLink(html, query string) (link string, err error) {
	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return
	}

	link, _ = doc.Find(query).First().Attr("href")

	return
}

func (a *Amazon) NormLink(amazonUrl url.URL, link string) (linkUrl *url.URL, err error) {
	linkUrl, err = url.Parse(link)
	if err != nil {
		return
	}

	if !strings.HasPrefix(link, "http") {
		linkUrl = amazonUrl.ResolveReference(linkUrl)
	}

	linkUrl.RawQuery = url.Values{}.Encode()

	return
}

func (a *Amazon) LoadAmazonPage(amazonUrl, referrerUrl url.URL) (html string, err error) {
	a.browser.SetHeader("referer", referrerUrl.String())

	userAgent, ok := a.userAgent.Next()
	if ok {
		a.browser.UserAgent(userAgent.Header())
	}

	var proxyAddr string
	proxyAddr, ok = a.proxies.Next()
	if ok {
		err = a.browser.Proxy(proxyAddr)
		if err != nil {
			return
		}
	}

	err = a.browser.Run(amazonUrl.String(), []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				var html string
				err = chromedp.OuterHTML("html", &html).Do(ctx)
				if err != nil {
					return
				}

				if strings.Contains(html, "ref=cs_503_link") {
					err = DetectedAutomation

					a.browser.Cancel()
				}
			}()

			return
		}),
		chromedp.OuterHTML("html", &html),
	})
	if err == DetectedAutomation {
		return a.LoadAmazonWithGoogle(amazonUrl)
	}

	return
}

func (a *Amazon) LoadAmazonWithGoogle(amazonUrl url.URL) (html string, err error) {
	userAgent, ok := a.userAgent.Next()
	if ok {
		a.browser.UserAgent(userAgent.Header())
	}

	var proxyAddr string
	proxyAddr, ok = a.proxies.Next()
	if ok {
		err = a.browser.Proxy(proxyAddr)
		if err != nil {
			return
		}
	}

	googleUrl := url.URL{
		Scheme: "https",
		Host:   "www.google.com",
		Path:   "search",
		RawQuery: url.Values{
			"q": []string{amazonUrl.String()},
		}.Encode(),
	}

	var result string
	err = a.browser.Run(googleUrl.String(), []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible("#search", chromedp.ByID),
		chromedp.OuterHTML("html", &result),
	})
	if err != nil {
		return
	}

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(strings.NewReader(result))
	if err != nil {
		return
	}

	var cachedUrl string
	doc.Find("#search a").EachWithBreak(func(i int, sel *goquery.Selection) (ok bool) {
		cachedUrl = sel.AttrOr("href", "")

		if strings.HasPrefix(cachedUrl, "https://webcache") {
			fmt.Println(cachedUrl)

			return true
		}

		return false
	})

	if len(cachedUrl) == 0 {
		return
	}

	a.browser.SetHeader("referer", googleUrl.String())

	err = a.browser.Run(cachedUrl, []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.OuterHTML("html", &html),
	})
	if err != nil {
		return
	}

	return
}

func (a *Amazon) extractFromSocPage(asin dao.ASIN, pagePath, linkQuery string) (meta Meta, ok bool, err error) {
	domains := a.domains
	sort.Slice(domains, func(i, j int) bool {
		return rand.Int31n(100) > 50
	})

	var (
		referrerUrl, amazonUrl url.URL
		html, link             string
		linkUrl                *url.URL
	)

	for _, domain := range domains {
		amazonUrl = url.URL{
			Scheme: "https",
			Host:   domain,
			Path:   fmt.Sprintf(pagePath, asin),
		}

		referrerUrl = url.URL{
			Scheme: "https",
			Host:   "www.google.com",
			Path:   "search",
			RawQuery: url.Values{
				"q": []string{string(asin)},
			}.Encode(),
		}

		html, err = a.LoadAmazonPage(amazonUrl, referrerUrl)
		if err != nil {
			log.WarnFmt("Can't load page `%s`: %v", amazonUrl.String(), err)

			continue
		}

		link, err = a.ExtractLink(html, linkQuery)
		if err != nil {
			log.WarnFmt("Can't extract the product URL `%s`: %v", amazonUrl.String(), err)

			continue
		}

		linkUrl, err = a.NormLink(amazonUrl, link)
		if err != nil {
			log.WarnFmt("Can't normalize the product URL `%s`: %v", link, err)

			continue
		}

		html, err = a.LoadAmazonPage(*linkUrl, amazonUrl)
		if err != nil {
			log.WarnFmt("Can't load page `%s`: %v", linkUrl.String(), err)

			continue
		}

		meta, err = a.ExtractMeta(html)
		if err != nil {
			log.WarnFmt("Get error in parsing processing of URL `%s` for ASIN `%s`: %v", amazonUrl.String(), asin, err)

			continue
		}

		ok = true

		break
	}

	return
}

func (a *Amazon) ExtractFromReview(asin dao.ASIN) (meta Meta, ok bool, err error) {
	return a.extractFromSocPage(asin, "/review/product/%v/", ".product-info a.a-link-normal[href]")
}

func (a *Amazon) ExtractFromQA(asin dao.ASIN) (meta Meta, ok bool, err error) {
	return a.extractFromSocPage(asin, "/ask/questions/asin/%v/", ".askProductDescription a.a-size-large.a-link-normal[href]")
}

func (a *Amazon) ExtractFromProduct(asin dao.ASIN) (meta Meta, ok bool, err error) {
	domains := a.domains
	sort.Slice(domains, func(i, j int) bool {
		return rand.Int31n(100) > 50
	})

	var (
		referrerUrl, amazonUrl url.URL
		html                   string
	)

	for _, domain := range domains {
		referrerUrl = url.URL{
			Scheme: "https",
			Host:   "www.google.com",
			Path:   "search",
			RawQuery: url.Values{
				"q": []string{string(asin)},
			}.Encode(),
		}

		amazonUrl = url.URL{
			Scheme: "https",
			Host:   domain,
			Path:   fmt.Sprintf("gp/product/%v/", asin),
		}

		html, err = a.LoadAmazonPage(amazonUrl, referrerUrl)
		if err != nil {
			continue
		}

		meta, err = a.ExtractMeta(html)
		if err != nil {
			log.WarnFmt("Get error in parsing processing of URL `%s` for ASIN `%s`: %v", amazonUrl.String(), asin, err)

			continue
		}

		ok = true

		break
	}

	return
}
