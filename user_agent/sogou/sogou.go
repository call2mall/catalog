package sogou

import (
	"sync"
)

type Sogou struct {
	mx   *sync.Mutex
	ix   int
	list []string
}

func NewSogou() (ua *Sogou) {
	return &Sogou{
		mx: &sync.Mutex{},
		ix: 0,
		list: []string{
			"Sogou Pic Spider/3.0( http://www.sogou.com/docs/help/webmasters.htm#07)",
			"Sogou head spider/3.0( http://www.sogou.com/docs/help/webmasters.htm#07)",
			"Sogou web spider/4.0(+http://www.sogou.com/docs/help/webmasters.htm#07)",
			"Sogou Orion spider/3.0( http://www.sogou.com/docs/help/webmasters.htm#07)",
			"Sogou-Test-Spider/4.0 (compatible; MSIE 5.5; Windows 98)",
		},
	}
}

func (ua Sogou) Header() (value string) {
	ua.mx.Lock()
	defer ua.mx.Unlock()

	ua.ix++
	if ua.ix >= len(ua.list) {
		ua.ix = 0
	}

	value = ua.list[ua.ix]

	return
}
