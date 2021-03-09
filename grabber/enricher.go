package grabber

import (
	"fmt"
	"github.com/call2mall/catalog/amazon"
	"github.com/call2mall/catalog/curl"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/proxy"
	"github.com/call2mall/catalog/translator"
	"github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"time"
)

func RunEnrichASIN(threads uint) (err error) {
	go runDefrostEnrichASIN()

	var ch chan dao.ASIN
	ch, err = runEnrichASINThreads(threads)
	if err != nil {
		return
	}

	var (
		asinList dao.ASINList
		asin     dao.ASIN
	)
	for {
		asinList, err = dao.PopASINToEnrich(threads)
		if err != nil {
			log.CriticalFmt("Can't pop new ASIN to enrich its pages: %v", err)

			continue
		}

		if len(asinList) == 0 {
			time.Sleep(time.Minute)

			continue
		}

		for _, asin = range asinList {
			ch <- asin
		}
	}
}

func runDefrostEnrichASIN() {
	const minute = 60

	var err error
	for range time.NewTicker(time.Minute).C {
		err = dao.DefrostEnrichASIN(minute)
		if err != nil {
			log.CriticalFmt("Can't defrost ASIN from enrich queue: %v", err)

			continue
		}
	}
}

func runEnrichASINThreads(threads uint) (ch chan dao.ASIN, err error) {
	ch = make(chan dao.ASIN, threads)

	var proxies *proxy.Proxies
	proxies, err = proxy.GetInstance()
	if err != nil {
		return
	}

	var i uint
	for i = 0; i < threads; i++ {
		go enrichByASIN(ch, proxies)
	}

	return
}

func enrichByASIN(ch chan dao.ASIN, proxies *proxy.Proxies) {
	var (
		err error

		ok bool

		originList dao.OriginList
		features   dao.ASINFeatures
	)

	for asin := range ch {
		log.DebugFmt("It is enriching data for ASIN `%s`", asin)

		originList, err = asin.LoadOrigins()
		if err != nil {
			return
		}

		features, ok, err = extractASINFeatures(asin, originList, proxies)
		if err != nil {
			err = fmt.Errorf("can't extract meta-data for ASIN `%s`: %s", asin, err.Error())

			log.Error(err.Error())

			continue
		}

		if !ok {
			err = asin.MarkEnrichAs(dao.Fail)
			if err != nil {
				err = fmt.Errorf("can't set status as `fail` of enricher queue task for ASIN `%s`: %s", asin, err.Error())

				log.Error(err.Error())

				continue
			}
		}

		err = features.Store()
		if err != nil {
			err = fmt.Errorf("can't store found ASIN features for ASIN `%s`: %s", asin, err.Error())

			e := asin.MarkEnrichAs(dao.Fail)
			if e != nil {
				err = errors.Wrap(err, e.Error())
			}

			log.Critical(err.Error())

			continue
		}

		err = asin.MarkEnrichAs(dao.Done)
		if err != nil {
			err = fmt.Errorf("can't set status as `done` of enricher queue task for ASIN `%s`: %s", asin, err.Error())

			log.Error(err.Error())

			continue
		}

		err = asin.PushToPublisherQueue()
		if err != nil {
			err = fmt.Errorf("can't push ASIN `%s` to queue to publish its: %s", asin, err.Error())

			log.Error(err.Error())

			continue
		}

		log.DebugFmt("Meta-data for ASIN `%s` is found", asin)
	}
}

func extractASINFeatures(asin dao.ASIN, originList dao.OriginList, proxies *proxy.Proxies) (features dao.ASINFeatures, ok bool, err error) {
	var (
		lang     string
		langList = []string{"en", "de", "it", "fr", "es", "nl", "sv", "ar", "jp"}

		pageUrl string

		meta  dao.ASINMeta
		image dao.Image
	)

	for _, lang = range langList {
		pageUrl, ok = originList[lang]
		if !ok {
			continue
		}

		meta, image, ok, err = extractASINMeta(pageUrl, proxies)
		if err != nil {
			log.DebugFmt("Can't enrich ASIN by page `%s`: %v", pageUrl, err)

			continue
		}

		if !ok {
			log.DebugFmt("Can't extract necessary features for ASIN `%s` by url `%s`", asin, pageUrl)

			continue
		}

		ok = false

		if lang != "en" {
			if len(meta.Title) == 0 {
				log.WarnFmt("Get features with empty title for ASIN `%s` by url `%s`", asin, pageUrl)

				continue
			}

			if len(meta.Category.Name) == 0 {
				log.WarnFmt("Get features with empty category name for ASIN `%s` by url `%s`", asin, pageUrl)

				continue
			}

			meta.Title, err = translator.Translate(meta.Title, lang, "en", proxies)
			if err != nil {
				log.ErrorFmt("Can't translate title with ASIN `%s`: %v", asin, err)

				continue
			}

			if len(meta.Title) == 0 {
				log.WarnFmt("Get translated empty title for ASIN `%s` by url `%s`", asin, pageUrl)

				continue
			}

			meta.Category.Name, err = translator.Translate(meta.Category.Name, lang, "en", proxies)
			if err != nil {
				log.ErrorFmt("Can't translate category for ASIN `%s`: %s", asin, err.Error())

				continue
			}

			if len(meta.Category.Name) == 0 {
				log.WarnFmt("Get translated empty category name for ASIN `%s` by url `%s`", asin, pageUrl)

				continue
			}
		}

		features.ASIN = asin
		features.Image = image
		features.ASINMeta = meta

		ok = true

		return
	}

	return
}

func extractASINMeta(pageUrl string, proxies *proxy.Proxies) (meta dao.ASINMeta, image dao.Image, ok bool, err error) {
	var features amazon.Features
	features, ok, err = amazon.ExtractFeaturesByUrl(pageUrl, proxies)
	if err != nil {
		err = fmt.Errorf("can't find amazon page for ASIN: %v", err)

		return
	}

	if !ok {
		return
	}

	if len(features.Category) == 0 || len(features.Title) == 0 {
		ok = false

		return
	}

	meta.Category.Name = features.Category
	meta.Title = features.Title

	var state uint
	image.Bytes, state, err = curl.LoadByUrl(features.PhotoUrl, header, proxies)
	if err != nil {
		err = fmt.Errorf("can't load image from `%s`: %v", pageUrl, err)

		return
	} else if state != 200 {
		err = fmt.Errorf("can't load image from `%s` because the server returned unexpected state", pageUrl)

		return
	}

	return
}
