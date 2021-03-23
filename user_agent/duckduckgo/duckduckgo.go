package duckduckgo

type DuckDuckGoBot struct {
}

func (ua DuckDuckGoBot) Header() (value string) {
	return "DuckDuckBot/1.0; (+http://duckduckgo.com/duckduckbot.html)"
}
