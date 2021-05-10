package search

import (
	"errors"
	"github.com/call2mall/catalog/chrome"
	"github.com/call2mall/catalog/proxy"
	"github.com/call2mall/catalog/search/bing"
	"github.com/call2mall/catalog/search/duckduckgo"
	"github.com/call2mall/catalog/search/google"
	"github.com/call2mall/catalog/search/lycos"
	"github.com/call2mall/catalog/search/qwant"
	"github.com/call2mall/catalog/search/swisscows"
	"github.com/call2mall/catalog/search/yahoo"
	"sync"
	"time"
)

type Searcher interface {
	Search(query string, b *chrome.Chrome) (urlList []string, err error)
}

func GetAllSearcherList() (list []Searcher) {
	return []Searcher{
		bing.Bing{},
		duckduckgo.DuckDuckGo{},
		lycos.Lycos{},
		qwant.Qwant{},
		swisscows.SwissCows{},
		yahoo.Yahoo{},
		google.Google{},
	}
}

type Search struct {
	mx      *sync.Mutex
	rotator *Rotator
	proxies *proxy.Proxies

	timeout time.Duration
}

func NewSearch(proxies *proxy.Proxies) (s *Search) {
	s = &Search{
		mx:      &sync.Mutex{},
		rotator: NewRotator(GetAllSearcherList()),
		proxies: proxies,
		timeout: time.Minute,
	}

	return
}

func (s *Search) SearcherList(list []Searcher) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.rotator = NewRotator(list)
}

func (s *Search) Timeout(timeout time.Duration) {
	s.timeout = timeout
}

var NonSearcher = errors.New("non searcher")

func (s *Search) Search(query string) (urlList []string, err error) {
	searcher, ok := s.rotator.Next()
	if !ok {
		err = NonSearcher

		return
	}

	var proxyAddr string
	proxyAddr, ok = s.proxies.Next()
	if !ok {
		err = proxy.NonProxy

		return
	}

	b := chrome.New()
	err = b.Proxy(proxyAddr)
	if err != nil {
		return
	}

	b.Timeout(s.timeout)

	urlList, err = searcher.Search(query, b)

	return
}
