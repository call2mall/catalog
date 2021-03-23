package alexa

type AlexaBot struct {
}

func (ua AlexaBot) Header() (value string) {
	return "ia_archiver (+http://www.alexa.com/site/help/webmasters; crawler@alexa.com)"
}
