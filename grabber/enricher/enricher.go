package enricher

import (
	"github.com/call2mall/catalog/amazon"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/proxy"
	"github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

var header = http.Header{}

func init() {
	header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
}

func RunEnricher(threads uint) (err error) {
	go defrostQueue()

	var ch chan dao.ASIN
	ch, err = runThreads(threads)
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

func defrostQueue() {
	const minute = 60

	var err error
	for range time.NewTicker(time.Minute).C {
		err = dao.DefrostEnrichQueue(minute)
		if err != nil {
			log.CriticalFmt("Can't defrost ASIN from enrich queue: %v", err)

			continue
		}
	}
}

func runThreads(threads uint) (ch chan dao.ASIN, err error) {
	ch = make(chan dao.ASIN, threads)

	var proxies *proxy.Proxies
	proxies, err = proxy.GetInstance()
	if err != nil {
		return
	}

	var i uint
	for i = 0; i < threads; i++ {
		go enrichProps(ch, proxies)
	}

	return
}

func enrichProps(ch chan dao.ASIN, proxies *proxy.Proxies) {
	var (
		err error

		lang string

		pageUrl    string
		ok         bool
		originList dao.OriginList

		a     amazon.Amazon
		props dao.ASINProps
	)

	for asin := range ch {
		func() {
			var success = false

			defer func() {
				if !success {
					e := asin.MarkEnrichAs(dao.Fail)
					if e != nil {
						err = errors.Wrap(err, e.Error())

						log.Warn(err.Error())
					}
				}
			}()

			log.DebugFmt("It is enriching data for ASIN `%s`", asin)

			originList, err = asin.LoadOrigins()
			if err != nil {
				return
			}

			for _, lang = range dao.CountryPriority() {
				pageUrl, ok = originList[lang]
				if !ok {
					continue
				}

				props, err = a.ExtractProps(pageUrl, proxies)
				if err != nil {
					log.ErrorFmt("Can't extract meta-data for ASIN `%s`: %s", asin, err.Error())

					continue
				}

				if len(props.Image.Bytes) > 0 {
					break
				}
			}

			if len(props.Image.Bytes) == 0 {
				log.WarnFmt("Can't extract meta-data for ASIN `%s` because nowhere is necessary data", asin)

				return
			}

			err = props.Store()
			if err != nil {
				log.CriticalFmt("Can't store found ASIN props for ASIN `%s`: %s", asin, err.Error())

				return
			}

			err = asin.MarkEnrichAs(dao.Done)
			if err != nil {
				log.ErrorFmt("Can't set status as `done` of enricher queue task for ASIN `%s`: %s", asin, err.Error())

				return
			}

			err = asin.PushToPublisherQueue()
			if err != nil {
				log.ErrorFmt("Can't push ASIN `%s` to queue to publish its: %s", asin, err.Error())

				return
			}

			log.DebugFmt("Meta-data for ASIN `%s` is found", asin)
		}()
	}
}
