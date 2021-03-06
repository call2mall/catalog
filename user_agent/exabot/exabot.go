package exabot

import (
	"sync"
)

type Exabot struct {
	mx   *sync.Mutex
	ix   int
	list []string
}

func NewExabot() (ua *Exabot) {
	return &Exabot{
		mx: &sync.Mutex{},
		ix: 0,
		list: []string{
			"Mozilla/5.0 (compatible; Konqueror/3.5; Linux) KHTML/3.5.5 (like Gecko) (Exabot-Thumbnails)",
			"Mozilla/5.0 (compatible; Exabot/3.0; +http://www.exabot.com/go/robot)",
		},
	}
}

func (ua Exabot) Header() (value string) {
	ua.mx.Lock()
	defer ua.mx.Unlock()

	ua.ix++
	if ua.ix >= len(ua.list) {
		ua.ix = 0
	}

	value = ua.list[ua.ix]

	return
}
