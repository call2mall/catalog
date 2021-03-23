package yahoo

type Yahoo struct {
}

func (ua Yahoo) Header() (value string) {
	return "Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp)"
}
