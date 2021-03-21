package searcher

import (
	"fmt"
	"github.com/call2mall/catalog/amazon"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/proxy"
	"github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"time"
)

func RunSearchASIN(threads uint) (err error) {
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
		asinList, err = dao.PopASINToSearch(threads)
		if err != nil {
			log.CriticalFmt("Can't pop new ASIN to search its pages: %v", err)

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
		err = dao.DefrostSearchASIN(minute)
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
					e := asin.MarkSearchAs(dao.Fail)
					if e != nil {
						err = errors.Wrap(err, e.Error())

						log.Warn(err.Error())
					}
				}
			}()

			urlList, err = a.FindPages(asin, proxies)
			if err != nil {
				err = fmt.Errorf("can't find origin amazon pages for ASIN `%s`: %v", asin, err)

				log.Warn(err.Error())

				return
			}

			originalList, err = dao.UrlsToOriginList(urlList)
			if err != nil {
				log.Warn(err.Error())

				return
			}

			if len(originalList) == 0 {
				err = fmt.Errorf("found zero amazon pages for ASIN `%s`", asin)

				log.Warn(err.Error())

				return
			}

			err = originalList.Store(asin)
			if err != nil {
				err = fmt.Errorf("can't store found amazon pages for ASIN `%s`: %s", asin, err.Error())

				log.Critical(err.Error())

				return
			}

			success = true

			err = asin.MarkSearchAs(dao.Done)
			if err != nil {
				err = fmt.Errorf("can't set status as `done` of searcher queue task for ASIN `%s`: %s", asin, err.Error())

				log.Error(err.Error())

				return
			}

			err = asin.PushToEnricherQueue()
			if err != nil {
				err = fmt.Errorf("can't push ASIN `%s` to queue to enrich meta-data: %s", asin, err.Error())

				log.Error(err.Error())

				return
			}

			log.DebugFmt("Page for ASIN `%s` is found", asin)
		}()
	}
}
