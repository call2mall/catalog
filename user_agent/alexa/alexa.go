package alexa

type Alexa struct {
}

func (ua Alexa) Header() (value string) {
	return "ia_archiver (+http://www.alexa.com/site/help/webmasters; crawler@alexa.com)"
}
