package amazon

import (
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/call2mall/catalog/browser"
	"github.com/call2mall/catalog/curl"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/proxy"
	"github.com/call2mall/catalog/search"
	"github.com/call2mall/catalog/search/lycos"
	"github.com/call2mall/catalog/search/qwant"
	"github.com/call2mall/catalog/search/swisscows"
	"github.com/call2mall/catalog/translate"
	"github.com/call2mall/catalog/user_agent"
	"github.com/chromedp/chromedp"
	. "github.com/pkg/errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Amazon struct{}

func (a Amazon) FindPages(asin dao.ASIN, proxies *proxy.Proxies) (urlList []string, err error) {
	s := search.NewSearch(proxies)
	s.SearcherList([]search.Searcher{
		lycos.Lycos{},
		qwant.Qwant{},
		swisscows.SwissCows{},
	})

	var (
		domainsRegExp = regexp.MustCompile("^(www\\.)?amazon\\.(co\\.uk|com\\.au|co\\.jp|com\\.mx|ae|sa|de|it|fr|es|nl|se|sg|in|ca)$")
		list          []string
	)

	list, err = s.Search(fmt.Sprintf("\"%s\"", asin))
	if err != nil {
		return
	}

	var (
		urlData *url.URL
		matches [][]string
	)

	b := browser.NewBrowser()
	proxyAddr, ok := proxies.Next()
	if ok {
		err = b.Proxy(proxyAddr)
		if err != nil {
			return
		}
	}

	for _, urlRaw := range list {
		urlData, err = url.Parse(urlRaw)
		if err != nil {
			return
		}

		matches = domainsRegExp.FindAllStringSubmatch(urlData.Host, -1)
		if len(matches) == 0 {
			continue
		}

		if strings.Contains(urlData.Path, "/review/product/") {
			urlRaw, ok, err = extractOriginFromProductReport(urlData, b)
			if err != nil || !ok {
				continue
			}
		} else if !strings.Contains(urlData.Path, "/dp/") {
			continue
		}

		urlList = append(urlList, urlRaw)
	}

	if len(urlList) == 0 {
		urlList, err = searchThroughProductReport(asin, b)
		if err != nil {
			return
		}
	}

	if len(urlList) == 0 {
		urlList, err = searchThroughProductQA(asin, b)
		if err != nil {
			return
		}
	}

	return
}

var domains = []string{
	"amazon.ca",
	"amazon.co.uk",
	"amazon.com.au",
	"amazon.co.jp",
	"amazon.sg",
	"amazon.nl",
	"amazon.se",
	"amazon.de",
	"amazon.fr",
	"amazon.es",
	"amazon.com.mx",
	"amazon.it",
	"amazon.ae",
	"amazon.sa",
	"amazon.in",
}

func searchThroughProductReport(asin dao.ASIN, b *browser.Browser) (urlList []string, err error) {
	var (
		urlData url.URL

		resUrl string
		ok     bool

		pos int
	)

	for _, host := range domains {
		urlData = url.URL{
			Scheme: "https",
			Host:   host,
			Path:   fmt.Sprintf("/review/product/%s", asin),
		}

		resUrl, ok, err = extractOriginFromProductReport(&urlData, b)
		if err != nil {
			return
		}

		if ok {
			pos = strings.Index(resUrl, "/ref=")
			if pos > -1 {
				resUrl = resUrl[0:pos]
			}

			urlList = append(urlList, resUrl)
		}
	}

	return
}

func extractOriginFromProductReport(urlData *url.URL, b *browser.Browser) (resUrl string, ok bool, err error) {
	var html string
	html, err = b.GetHtml(urlData.String())
	if err != nil {
		return
	}

	reader := strings.NewReader(html)

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		return
	}

	resUrl, ok = doc.Find(".product-info a.a-link-normal[href]").First().Attr("href")

	var resData *url.URL
	resData, err = url.Parse(resUrl)
	if err != nil {
		return
	}

	if !strings.HasPrefix(resUrl, "http") {
		resData = urlData.ResolveReference(resData)
	}

	resData.RawQuery = url.Values{}.Encode()

	resUrl = resData.String()

	return
}

func searchThroughProductQA(asin dao.ASIN, b *browser.Browser) (urlList []string, err error) {
	var (
		urlData url.URL

		resUrl string
		ok     bool

		pos int
	)

	for _, host := range domains {
		urlData = url.URL{
			Scheme: "https",
			Host:   host,
			Path:   fmt.Sprintf("/ask/questions/asin/%s", asin),
		}

		resUrl, ok, err = extractOriginFromAQ(&urlData, b)
		if err != nil {
			return
		}

		if ok {
			pos = strings.Index(resUrl, "/ref=")
			if pos > -1 {
				resUrl = resUrl[0:pos]
			}

			urlList = append(urlList, resUrl)
		}
	}

	return
}

func extractOriginFromAQ(urlData *url.URL, b *browser.Browser) (resUrl string, ok bool, err error) {
	var html string
	html, err = b.GetHtml(urlData.String())
	if err != nil {
		return
	}

	reader := strings.NewReader(html)

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		return
	}

	resUrl, ok = doc.Find(".askProductDescription a.a-size-large.a-link-normal[href]").First().Attr("href")

	var resData *url.URL
	resData, err = url.Parse(resUrl)
	if err != nil {
		return
	}

	if !strings.HasPrefix(resUrl, "http") {
		resData = urlData.ResolveReference(resData)
	}

	resData.RawQuery = url.Values{}.Encode()

	resUrl = resData.String()

	return
}

var (
	NoProductCategory  = errors.New("there isn't product category")
	NoProductTitle     = errors.New("there isn't product title")
	NoProductImage     = errors.New("there isn't product image")
	DetectedAutomation = errors.New("detected automation")
)

func (a Amazon) ExtractProps(amazonUrl string, proxies *proxy.Proxies) (props dao.ASINProps, err error) {
	var (
		l8n           string
		withL8nSwitch bool
		withTranslate bool

		detectedL8n translate.L8n
	)
	_, l8n, withL8nSwitch, err = translate.DetectL8nByAmazonUrl(amazonUrl)
	if err != nil {
		return
	}

	withTranslate = l8n != "en"

	var urlData *url.URL
	urlData, err = url.Parse(amazonUrl)
	if err != nil {
		return
	}

	if withL8nSwitch {
		urlData.RawQuery = url.Values{
			"language": []string{"en"},
		}.Encode()
	}

	proxyAddr, ok := proxies.Next()
	if !ok {
		err = proxy.NonProxy

		return
	}

	b := browser.NewBrowser()
	err = b.Proxy(proxyAddr)
	if err != nil {
		return
	}

	var ua user_agent.UserAgent
	ua, ok = user_agent.GetInstance().Next()
	if ok {
		b.UserAgent(ua.Header())
	}

	var (
		html                 string
		isAutomationDetected bool
	)
	err = b.Run(amazonUrl, []chromedp.Action{
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				var html string
				err = chromedp.OuterHTML("html", &html).Do(ctx)
				if err != nil {
					return
				}

				if strings.Contains(html, "ref=cs_503_link") {
					isAutomationDetected = true

					b.Cancel()
				}
			}()

			return
		}),
		chromedp.Sleep(time.Second),
		chromedp.OuterHTML("html", &html),
	})

	if isAutomationDetected {
		html, err = extractPropByGoogleCache(amazonUrl, proxies)
		if err != nil {
			err = Wrap(err, fmt.Sprintf("amazon web-site `%s` responses: %v", urlData.Host, DetectedAutomation))

			return
		}
	}

	if err != nil {
		return
	}

	reader := strings.NewReader(html)

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(reader)
	if err != nil {
		return
	}

	props.Title = doc.Find("#title span").First().Text()
	props.Title = strings.TrimSpace(props.Title)
	props.Title = strings.Replace(props.Title, " ", "", -1)
	if len(props.Title) == 0 {
		err = NoProductTitle

		return
	}

	if !withTranslate {
		detectedL8n = translate.DetectL8n(props.Title)
		if detectedL8n != translate.EnglishL8n {
			withTranslate = true

			l8n = string(detectedL8n)
		}
	}

	if withTranslate {
		tr := translate.NewTranslate(proxies)
		props.Title, err = tr.Translate(props.Title, l8n, "en")
		if err != nil {
			return
		}

		if len(props.Title) == 0 {
			err = NoProductTitle

			return
		}
	}

	doc.Find("#wayfinding-breadcrumbs_feature_div ul li a").Each(func(i int, sel *goquery.Selection) {
		var category = sel.Text()
		category = strings.TrimSpace(category)
		category = strings.ReplaceAll(category, "&amp;", "and")

		fmt.Println(category)
	})

	props.Category.Name = doc.Find("#wayfinding-breadcrumbs_feature_div ul li:first-of-type a").First().Text()
	props.Category.Name = strings.TrimSpace(props.Category.Name)
	props.Category.Name = strings.ReplaceAll(props.Category.Name, "&amp;", "and")
	if len(props.Category.Name) == 0 {
		err = NoProductCategory

		return
	} else if !withTranslate {
		detectedL8n = translate.DetectL8n(props.Category.Name)
		if detectedL8n != translate.EnglishL8n {
			withTranslate = true

			l8n = string(detectedL8n)
		}
	}

	if withTranslate {
		tr := translate.NewTranslate(proxies)
		props.Category.Name, err = tr.Translate(props.Category.Name, l8n, "en")
		if err != nil {
			return
		}

		if len(props.Category.Name) == 0 {
			err = NoProductCategory

			return
		}
	}

	if len(props.Category.Name) > 1 {
		props.Category.Name = strings.ToUpper(props.Category.Name[0:1]) + props.Category.Name[1:]
	}

	var photoUrl string
	photoUrl, ok = doc.Find("#imgTagWrapperId img[src]").First().Attr("src")
	if !ok {
		err = NoProductImage

		return
	}

	header := http.Header{}
	header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")

	var state uint
	props.Image.Bytes, state, err = curl.LoadByUrl(photoUrl, header, proxies)
	if err != nil {
		err = fmt.Errorf("can't load image from `%s`: %v", amazonUrl, err)

		return
	} else if state != 200 {
		err = fmt.Errorf("can't load image from `%s` because the server returned unexpected state", amazonUrl)

		return
	}

	return
}

func extractPropByGoogleCache(amazonUrl string, proxies *proxy.Proxies) (html string, err error) {
	b := browser.NewBrowser()
	proxyAddr, ok := proxies.Next()
	if ok {
		err = b.Proxy(proxyAddr)
		if err != nil {
			return
		}
	}

	queryUrl := url.URL{
		Scheme: "https",
		Host:   "www.google.com",
		Path:   "search",
		RawQuery: url.Values{
			"q": []string{amazonUrl},
		}.Encode(),
	}

	var result string
	err = b.Run(queryUrl.String(), []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible("#search", chromedp.ByID),
		chromedp.OuterHTML("html", &result),
	})
	if err != nil {
		return
	}

	reader := strings.NewReader(result)

	var doc *goquery.Document
	doc, err = goquery.NewDocumentFromReader(reader)
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

	b.SetHeader("referer", queryUrl.String())

	err = b.Run(cachedUrl, []chromedp.Action{
		chromedp.Sleep(time.Second),
		chromedp.OuterHTML("html", &html),
	})
	if err != nil {
		return
	}

	return
}
