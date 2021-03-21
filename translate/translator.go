package translate

import (
	"errors"
	"github.com/call2mall/catalog/proxy"
	"github.com/call2mall/catalog/translate/translate"
	"github.com/call2mall/catalog/translate/webtran"
	"sync"
	"time"
)

type Translator interface {
	Translate(text, from, to, proxyAddr string) (result string, err error)
	Timestamp(tm time.Duration)
}

func GetAllTranslatorList() (list []Translator) {
	return []Translator{
		&translate.Translate{},
		&webtran.WebTran{},
	}
}

type Translate struct {
	mx      *sync.Mutex
	rotator *Rotator
	proxies *proxy.Proxies

	timeout time.Duration
}

func NewTranslate(proxies *proxy.Proxies) (s *Translate) {
	s = &Translate{
		mx:      &sync.Mutex{},
		rotator: NewRotator(GetAllTranslatorList()),
		proxies: proxies,
		timeout: time.Minute,
	}

	return
}

func (s *Translate) TranslatorList(list []Translator) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.rotator = NewRotator(list)
}

func (s *Translate) Timeout(timeout time.Duration) {
	s.timeout = timeout
}

var NonTranslator = errors.New("non translator")

func (s *Translate) Translate(text, from, to string) (result string, err error) {
	translator, ok := s.rotator.Next()
	if !ok {
		err = NonTranslator

		return
	}

	var proxyAddr string
	proxyAddr, ok = s.proxies.Next()
	if !ok {
		err = proxy.NonProxy

		return
	}

	result, err = translator.Translate(text, from, to, proxyAddr)

	return
}
