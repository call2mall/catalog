package yahoo

type YahooBot struct {
}

func (ua YahooBot) Header() (value string) {
	return "Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)"
}
