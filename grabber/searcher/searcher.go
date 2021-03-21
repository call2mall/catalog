package searcher

import (
	"fmt"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/google"
	"github.com/call2mall/catalog/proxy"
	"github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"time"
)

func RunSearchASIN(threads uint) (err error) {
	go runDefrostSearchASIN()

	var ch chan dao.ASIN
	ch, err = runSearchASINThreads(threads)
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

func runDefrostSearchASIN() {
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

func runSearchASINThreads(threads uint) (ch chan dao.ASIN, err error) {
	ch = make(chan dao.ASIN, threads)

	var proxies *proxy.Proxies
	proxies, err = proxy.GetInstance()
	if err != nil {
		return
	}

	var i uint
	for i = 0; i < threads; i++ {
		go searchOriginsByASIN(ch, proxies)
	}

	return
}

func searchOriginsByASIN(ch chan dao.ASIN, proxies *proxy.Proxies) {
	var (
		urlList      []string
		originalList dao.OriginList
		err          error
	)

	for asin := range ch {
		log.DebugFmt("It is searching page by ASIN `%s`", asin)

		urlList, err = google.FindPageByASIN(string(asin), proxies)
		if err != nil {
			err = fmt.Errorf("can't find amazon pages through google by ASIN `%s`: %v", asin, err)

			e := asin.MarkSearchAs(dao.Fail)
			if e != nil {
				err = errors.Wrap(err, e.Error())
			}

			log.Warn(err.Error())

			continue
		}

		if len(urlList) == 0 {
			urlList, err = SearchByASIN(string(asin), proxies)
			if err != nil {
				err = fmt.Errorf("can't find amazon pages through amazon review pages by ASIN `%s`: %v", asin, err)

				e := asin.MarkSearchAs(dao.Fail)
				if e != nil {
					err = errors.Wrap(err, e.Error())
				}

				log.Warn(err.Error())

				continue
			}
		}

		originalList, err = dao.ListToOriginList(urlList)
		if err != nil {
			e := asin.MarkSearchAs(dao.Fail)
			if e != nil {
				err = errors.Wrap(err, e.Error())
			}

			log.Warn(err.Error())

			continue
		}

		if len(originalList) == 0 {
			err = fmt.Errorf("found zero amazon pages for ASIN `%s`", asin)

			e := asin.MarkSearchAs(dao.Fail)
			if e != nil {
				err = errors.Wrap(err, e.Error())
			}

			log.Warn(err.Error())

			continue
		}

		err = originalList.Store(asin)
		if err != nil {
			err = fmt.Errorf("can't store found amazon pages for ASIN `%s`: %s", asin, err.Error())

			e := asin.MarkSearchAs(dao.Fail)
			if e != nil {
				err = errors.Wrap(err, e.Error())
			}

			log.Critical(err.Error())

			continue
		}

		err = asin.MarkSearchAs(dao.Done)
		if err != nil {
			err = fmt.Errorf("can't set status as `done` of searcher queue task for ASIN `%s`: %s", asin, err.Error())

			log.Error(err.Error())

			continue
		}

		err = asin.PushToEnricherQueue()
		if err != nil {
			err = fmt.Errorf("can't push ASIN `%s` to queue to enrich meta-data: %s", asin, err.Error())

			log.Error(err.Error())

			continue
		}

		log.DebugFmt("Page for ASIN `%s` is found", asin)
	}
}
