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
	"github.com/call2mall/catalog/translate"
	"github.com/chromedp/chromedp"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Amazon struct{}

var domainsRegExp = regexp.MustCompile("^(www\\.)?amazon\\.(co\\.uk|com\\.au|co\\.jp|ae|de|it|fr|es|nl|se|sg)$")

func (a Amazon) FindPages(asin dao.ASIN, proxies *proxy.Proxies) (urlList []string, err error) {
	s := search.NewSearch(proxies)

	var list []string
	list, err = s.Search(fmt.Sprintf("\"%s\"", asin))
	if err != nil {
		return
	}

	var (
		urlData *url.URL
		matches [][]string
	)

	for _, urlRaw := range list {
		urlData, err = url.Parse(urlRaw)
		if err != nil {
			return
		}

		matches = domainsRegExp.FindAllStringSubmatch(urlData.Host, -1)
		if len(matches) == 0 {
			continue
		}

		if !strings.Contains(urlData.Path, "/dp/") {
			continue
		}

		urlList = append(urlList, urlRaw)
	}

	return
}

var (
	NoProductCategory  = errors.New("there isn't product category")
	NoProductTitle     = errors.New("there isn't product title")
	NoProductImage     = errors.New("there isn't product image")
	DetectedAutomation = errors.New("detected automation")
)

func (a Amazon) ExtractMeta(amazonUrl string, proxies *proxy.Proxies) (props dao.ASINProps, err error) {
	var (
		l8n           string
		withL8nSwitch bool
		withTranslate bool
	)
	_, l8n, withL8nSwitch, err = dao.DetectL8nOfOrigin(amazonUrl)
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
		chromedp.WaitVisible("#sp-cc-accept", chromedp.ByID),
		chromedp.Sleep(time.Second),
		chromedp.Click("#sp-cc-accept", chromedp.ByID),
		chromedp.OuterHTML("html", &html),
	})

	if isAutomationDetected {
		err = DetectedAutomation

		return
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

	props.Category.Name = doc.Find("#wayfinding-breadcrumbs_feature_div ul li:first-of-type a").First().Text()
	props.Category.Name = strings.TrimSpace(props.Category.Name)
	if len(props.Category.Name) == 0 {
		err = NoProductCategory

		return
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
