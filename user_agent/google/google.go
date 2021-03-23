package google

import (
	"sync"
)

type GoogleBot struct {
	mx   *sync.Mutex
	ix   int
	list []string
}

func NewGoogleBot() (ua *GoogleBot) {
	return &GoogleBot{
		mx: &sync.Mutex{},
		ix: 0,
		list: []string{
			"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			"Mozilla/5.0 (compatible; Googlebot-News; +http://www.google.com/bot.html)",
			"Mozilla/5.0 (compatible; Googlebot-Image/1.0; +http://www.google.com/bot.html)",
			"Mozilla/5.0 (compatible; Googlebot-Video/1.0; +http://www.google.com/bot.html)",
			"SAMSUNG-SGH-E250/1.0 Profile/MIDP-2.0 Configuration/CLDC-1.1 UP.Browser/6.2.3.3.c.1.101 (GUI) MMP/2.0 (compatible; Googlebot-Mobile/2.1; +http://www.google.com/bot.html)",
			"Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2272.96 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
			"Mozilla/5.0 (compatible; Mediapartners-Google/2.1; +http://www.google.com/bot.html)",
			"Mediapartners-Google",
			"AdsBot-Google (+http://www.google.com/adsbot.html)",
			"AdsBot-Google-Mobile-Apps",
		},
	}
}

func (ua GoogleBot) Header() (value string) {
	ua.mx.Lock()
	defer ua.mx.Unlock()

	ua.ix++
	if ua.ix >= len(ua.list) {
		ua.ix = 0
	}

	value = ua.list[ua.ix]

	return
}
