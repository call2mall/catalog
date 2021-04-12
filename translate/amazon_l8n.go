package translate

import (
	"fmt"
	"net/url"
	"strings"
)

func DetectL8nByAmazonUrl(rawUrl string) (country, l8n string, withL8nSwitch bool, err error) {
	var urlData *url.URL
	urlData, err = url.Parse(rawUrl)
	if err != nil {
		return
	}

	urlData.Host = strings.Replace(urlData.Host, "www.", "", 1)
	urlData.Host = strings.Replace(urlData.Host, "amazon.", "", 1)

	switch urlData.Host {
	case "co.uk":
		country = "uk"
		l8n = "en"
	case "com.au":
		country = "au"
		l8n = "en"
	case "com.mx":
		country = "mx"
		l8n = "sp"
	case "ca":
		country = "ca"
		l8n = "en"
	case "sg":
		country = "sg"
		l8n = "en"
	case "ae":
		country = "ae"
		l8n = "en"
		withL8nSwitch = true
	case "sa":
		country = "sa"
		l8n = "en"
		withL8nSwitch = true
	case "co.jp":
		country = "jp"
		l8n = "en"
		withL8nSwitch = true
	case "de":
		country = "de"
		l8n = "en"
		withL8nSwitch = true
	case "it":
		country = "it"
		l8n = "it"
	case "fr":
		country = "fr"
		l8n = "fr"
	case "es":
		country = "es"
		l8n = "es"
	case "nl":
		country = "nl"
		l8n = "en"
		withL8nSwitch = true
	case "se":
		country = "se"
		l8n = "sv"
	case "in":
		country = "in"
		l8n = "hi"
		withL8nSwitch = true
	default:
		err = fmt.Errorf("can't detect localisation of `%s`", rawUrl)

		return
	}

	return
}
