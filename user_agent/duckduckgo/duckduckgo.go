package duckduckgo

type DuckDuckGo struct {
}

func (ua DuckDuckGo) Header() (value string) {
	return "DuckDuckBot/1.0; (+http://duckduckgo.com/duckduckbot.html)"
}
