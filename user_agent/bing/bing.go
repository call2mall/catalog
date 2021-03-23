package bing

type BingBot struct {
}

func (ua BingBot) Header() (value string) {
	return "Mozilla/5.0 (compatible; Bingbot/2.0; +http://www.bing.com/bingbot.htm)"
}
