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
		wevikabot.WeViKa{},
		google.NewGoogle(),
		duckduckgo.DuckDuckGo{},
		bing.Bing{},
		yahoo.Yahoo{},
		baidu.Baidu{},
		sogou.NewSogou(),
		yandex.Yandex{},
		exabot.NewExabot(),
		facebook.NewFacebook(),
		alexa.Alexa{},
	}
}
