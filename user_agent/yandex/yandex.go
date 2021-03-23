package yandex

type Yandex struct {
}

func (ua Yandex) Header() (value string) {
	return "Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)"
}
