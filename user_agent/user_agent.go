package user_agent

import (
	"github.com/call2mall/catalog/user_agent/alexa"
	"github.com/call2mall/catalog/user_agent/baidu"
	"github.com/call2mall/catalog/user_agent/bing"
	"github.com/call2mall/catalog/user_agent/duckduckgo"
	"github.com/call2mall/catalog/user_agent/exabot"
	"github.com/call2mall/catalog/user_agent/facebook"
	"github.com/call2mall/catalog/user_agent/google"
	"github.com/call2mall/catalog/user_agent/sogou"
	"github.com/call2mall/catalog/user_agent/wevikabot"
	"github.com/call2mall/catalog/user_agent/yahoo"
	"github.com/call2mall/catalog/user_agent/yandex"
)

type UserAgent interface {
	Header() (value string)
}

func GetAllUserAgentList() (list []UserAgent) {
	return []UserAgent{
		wevikabot.WeViKaBot{},
		google.NewGoogleBot(),
		duckduckgo.DuckDuckGoBot{},
		bing.BingBot{},
		yahoo.YahooBot{},
		baidu.BaiduBot{},
		sogou.NewSogouBot(),
		yandex.YandexBot{},
		exabot.NewExaBot(),
		facebook.NewFacebookBot(),
		alexa.AlexaBot{},
	}
}
