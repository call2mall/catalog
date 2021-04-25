package amazon

import (
	"fmt"
	"github.com/call2mall/catalog/browser"
	"github.com/call2mall/catalog/crawler"
	"github.com/call2mall/catalog/dao"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	config "github.com/leprosus/golang-config"
	log "github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	domains = []string{
		"amazon.com",
		"amazon.ca",
		"amazon.co.uk",
		"amazon.com.au",
		"amazon.sg",
	}
	paths = []string{
		"/review/product/%v/",
		"/ask/questions/asin/%v/",
	}
)

type Amazon struct {
	timeout time.Duration
}

type Meta struct {
	Url      string
	Title    string
	Category []string
	Bytes    []byte
}

func (m Meta) String() string {
	return fmt.Sprintf("{%v %v %v}", m.Title, m.Category, m.Url)
}

type UrlPair struct {
	PageUrl     string
	ReferrerUrl string
}

var (
	NoProduct         = errors.New("there isn't product")
	NoProductCategory = errors.New("there isn't product category")
	NoProductTitle    = errors.New("there isn't product title")
	NoProductImage    = errors.New("there isn't product image")
)

func NewAmazon() (a *Amazon) {
	a = &Amazon{
		timeout: time.Minute,
	}

	return
}

func (a *Amazon) Extract(asin dao.ASIN) (meta Meta, ok bool, err error) {
	for _, path := range paths {
		meta = a.extractFromSocPage(asin, path)

		ok = len(meta.Bytes) != 0
		if ok {
			return
		}
	}

	/*var (
		pair UrlPair
		pairList []UrlPair
	)
	for _, domain := range domains {
		pair.PageUrl = url.URL{
			Scheme: "https",
			Host:   domain,
			Path:   fmt.Sprintf("gp/product/%v/", asin),
		}

		pair.ReferrerUrl = url.URL{
			Scheme: "https",
			Host:   "www.google.com",
			Path:   "search",
			RawQuery: url.Values{
				"q": []string{string(asin)},
			}.Encode(),
		}

		pairList = append(pairList, pair)
	}

	meta, ok, err = a.extractMeta(pairList)
	if err != nil {
		return
	}

	if ok {
		return
	}*/

	return
}

func (a *Amazon) extractFromSocPage(asin dao.ASIN, path string) (meta Meta) {
	var (
		out  = make(chan Meta)
		done = make(chan interface{})
		wg   = &sync.WaitGroup{}
	)

	b := browser.NewBrowser()
	defer b.Close()
	b.Headless(config.BoolOrDefault("grabber.headless", true))

	go func() {
		meta = <-out
		done <- nil
	}()

	go func() {
		time.Sleep(a.timeout)
		done <- nil
	}()

	go func() {
		wg.Wait()
		done <- nil
	}()

	//TODO temp
	//b.WithTrace(true)
	//b.WithDevTools(true)

	var amazonUrl, referrerUrl url.URL

	referrerUrl = url.URL{
		Scheme: "https",
		Host:   "www.google.com",
		Path:   "search",
		RawQuery: url.Values{
			"q": []string{string(asin)},
		}.Encode(),
	}

	for _, domain := range domains {
		wg.Add(1)

		amazonUrl = url.URL{
			Scheme: "https",
			Host:   domain,
			Path:   fmt.Sprintf(path, asin),
		}

		go func(pair UrlPair) {
			var meta Meta

			defer wg.Done()

			err := b.Run(func(t *browser.Tab) (err error) {
				err = t.SetHeader("Referer", pair.ReferrerUrl)
				if err != nil {
					return
				}

				page := t.Page().Timeout(a.timeout)

				var done = make(chan bool)
				go page.EachEvent(func(e *proto.NetworkResponseReceived) {
					if e.Response.MIMEType == "text/html" {
						isTargetPage := strings.Contains(e.Response.URL, "/review/product/") ||
							strings.Contains(e.Response.URL, "/ask/questions/") ||
							strings.Contains(e.Response.URL, "/dp/") ||
							strings.Contains(e.Response.URL, "/gp/product/") ||
							strings.Contains(e.Response.URL, "/gp/errors/")
						if isTargetPage {
							done <- e.Response.Status == 200
						}
					}
				})()

				err = page.Navigate(pair.PageUrl)
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pair.PageUrl, err)

					return
				}

				isValid := <-done
				if !isValid {
					err = NoProduct

					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pair.PageUrl, err)

					return
				}

				var linkEl *rod.Element
				linkEl, err = page.Element(".askProductDescription .a-link-normal, .product-info .a-link-normal")
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pair.PageUrl, err)

					return
				}

				var linkUrl *string
				linkUrl, err = linkEl.Attribute("href")
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pair.PageUrl, err)

					return
				}

				var pageData url.URL
				pageData, err = normLink(pair.PageUrl, *linkUrl)
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pair.PageUrl, err)

					return
				}

				err = page.Navigate(pageData.String())
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				isValid = <-done
				if !isValid {
					err = NoProduct

					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				meta.Url = pageData.String()

				var titleEl *rod.Element
				titleEl, err = page.Element("#title span")
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				meta.Title, err = titleEl.Text()
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				meta.Title = strings.TrimSpace(meta.Title)
				meta.Title = strings.Replace(meta.Title, " ", "", -1)
				if len(meta.Title) == 0 {
					err = NoProductTitle

					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				var categoriesEls rod.Elements
				categoriesEls, err = page.Elements("#wayfinding-breadcrumbs_feature_div ul li a")
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				var name string
				for _, category := range categoriesEls {
					name, err = category.Text()
					if err != nil {
						log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

						return
					}

					name = strings.TrimSpace(name)
					name = strings.ReplaceAll(name, "&amp;", "and")

					meta.Category = append(meta.Category, name)
				}

				if len(meta.Category) == 0 {
					err = NoProductCategory

					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				var imageEl *rod.Element
				imageEl, err = page.Element("#imgTagWrapperId img[src][data-old-hires]")
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				var imageSrc *string
				imageSrc, err = imageEl.Attribute("src")
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				if len(*imageSrc) == 0 || strings.Contains(*imageSrc, "base64") {
					imageSrc, err = imageEl.Attribute("data-old-hires")
					if err != nil {
						log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

						return
					}
				}

				var imageUrl string
				imageUrl = *imageSrc

				if len(imageUrl) == 0 {
					err = NoProductImage

					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, pageData.String(), err)

					return
				}

				c := crawler.NewCrawler()
				err = c.Do(http.MethodGet, imageUrl, nil)
				if err != nil {
					log.DebugFmt("Loading image for ASIN %v on `%s` by url `%s`: %v", asin, pageData.String(), imageUrl, err)

					return
				}

				status := c.GetStatus()
				if status != 200 {
					err = NoProductImage

					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, imageUrl, err)

					return
				}

				meta.Bytes, err = io.ReadAll(c.GetBody())
				if err != nil {
					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, imageUrl, err)

					return
				}
				defer func() {
					_ = c.GetBody().Close()
				}()

				if len(meta.Bytes) == 0 {
					err = NoProductImage

					log.DebugFmt("Parsing ASIN %v on `%s`: %v", asin, *imageSrc, err)

					return
				}

				return
			})
			if err != nil {
				return
			}

			out <- meta
		}(UrlPair{
			PageUrl:     amazonUrl.String(),
			ReferrerUrl: referrerUrl.String(),
		})
	}

	<-done

	return
}

func normLink(amazonUrl string, link string) (linkUrl url.URL, err error) {
	var amazonData *url.URL
	amazonData, err = url.Parse(amazonUrl)
	if err != nil {
		return
	}

	var rawUrl *url.URL
	rawUrl, err = url.Parse(link)
	if err != nil {
		return
	}

	if !strings.HasPrefix(link, "http") {
		rawUrl = amazonData.ResolveReference(rawUrl)
	}

	rawUrl.RawQuery = url.Values{}.Encode()

	linkUrl = *rawUrl

	return
}
