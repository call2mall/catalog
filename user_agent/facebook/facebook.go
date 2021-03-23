package facebook

import (
	"sync"
)

type FacebookBot struct {
	mx   *sync.Mutex
	ix   int
	list []string
}

func NewFacebookBot() (ua *FacebookBot) {
	return &FacebookBot{
		mx: &sync.Mutex{},
		ix: 0,
		list: []string{
			"facebot",
			"facebookexternalhit/1.0 (+http://www.facebook.com/externalhit_uatext.php)",
			"facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
		},
	}
}

func (ua FacebookBot) Header() (value string) {
	ua.mx.Lock()
	defer ua.mx.Unlock()

	ua.ix++
	if ua.ix >= len(ua.list) {
		ua.ix = 0
	}

	value = ua.list[ua.ix]

	return
}
