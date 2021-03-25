package searcher

import (
	"github.com/call2mall/catalog/amazon"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/proxy"
	"github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"time"
)

func RunSearcher(threads uint) (err error) {
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
		asinList, err = dao.PopFromSearcher(threads)
		if err != nil {
			log.CriticalFmt("Can't pop new ASIN to searcher its pages: %v", err)

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
		err = dao.DefrostSearcher(minute)
		if err != nil {
			log.CriticalFmt("Can't defrost ASIN from search queue: %v", err)

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
		go searchOrigins(ch, proxies)
	}

	return
}

func searchOrigins(ch chan dao.ASIN, proxies *proxy.Proxies) {
	var (
		urlList      []string
		originalList dao.OriginList
		err          error

		a = amazon.Amazon{}
	)

	for asin := range ch {
		func() {
			var success = false

			defer func() {
				if !success {
					e := asin.MarkSearcherAs(dao.Fail)
					if e != nil {
						err = errors.Wrap(err, e.Error())

						log.Warn(err.Error())
					}
				}
			}()

			log.DebugFmt("It is searching page for ASIN `%s`", asin)

			urlList, err = a.FindPages(asin, proxies)
			if err != nil {
				log.WarnFmt("Can't find origin amazon pages for ASIN `%s`: %v", asin, err)

				return
			}

			originalList, err = dao.UrlsToOriginList(urlList)
			if err != nil {
				log.Warn(err.Error())

				return
			}

			if len(originalList) == 0 {
				log.WarnFmt("Found zero amazon pages for ASIN `%s`", asin)

				return
			}

			err = originalList.Store(asin)
			if err != nil {
				log.CriticalFmt("Can't store found amazon pages for ASIN `%s`: %v", asin, err)

				return
			}

			success = true

			err = asin.MarkSearcherAs(dao.Done)
			if err != nil {
				log.ErrorFmt("Can't set status as `done` of searcher queue task for ASIN `%s`: %v", asin, err)

				return
			}

			err = asin.PushToEnricher()
			if err != nil {
				log.ErrorFmt("Can't push ASIN `%s` to queue to enrich meta-data: %v", asin, err)

				return
			}

			log.DebugFmt("Page for ASIN `%s` is found", asin)
		}()
	}
}
