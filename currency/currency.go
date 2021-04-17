package currency

import (
	"encoding/json"
	"fmt"
	"github.com/call2mall/catalog/crawler"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Currency struct {
	Nominal uint    `json:"Nominal"`
	Value   float64 `json:"Value"`
}

type List struct {
	Currencies map[string]Currency `json:"Valute"`
}

var (
	refreshPeriod = 12 * time.Hour
	lastTime      = time.Now().Add(-10 * time.Second)
	currencies    sync.Map
)

func GetFactor(from, to string) (factor float64, err error) {
	if time.Now().After(lastTime) {
		lastTime = time.Now().Add(refreshPeriod)

		c := crawler.NewCrawler()

		header := http.Header{}
		header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
		header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
		header.Add("accept-language", "en-US;q=0.9,en;q=0.8")
		header.Add("cookie", "CGIC=Ij90ZXh0L2h0bWwsYXBwbGljYXRpb24veGh0bWwreG1sLGFwcGxpY2F0aW9uL3htbDtxPTAuOSwqLyo7cT0wLjg; SIDCC=ABtHo-H24FfZrVOAgbs2idHAq4bGxodmPfKpX4ItQgJ5E2j9auVGYxuNiQ1fI3N1Ot9dGE9TrNI; 1P_JAR=2018-12-3-15; DV=o5Cj-I74cwNRsO9g7sjbrkcaO5RJdxarpG_Mo3dUZgAAAGCtLueKQ3moKgAAAGC9VCFbQNaLNwAAAIQQ4-QFXGMHGAAAAA; GOOGLE_ABUSE_EXEMPTION=ID=c9a01c96fd993c95:TM=1543848057:C=r:IP=94.25.231.158-:S=APGng0tKXOkoBQTAJy7LyALNecO9DGZZqA; NID=148=Wo8GAudG0_Ol1KsXgNqhJcCCv00qOoroDNnAl_fJdZvIsfCH3L9DesZLfxlltHfYv5qPx7J6URy40-kpBRQUWWIMpDGs5HZAnd8xApt_j2_QWh72a_HaoXz7uJu55y1Yofw6uiCMhsIqfvBJpzZzcBYerODDh3T5TvNo2AiwvI3I1CcW4Vl7HBKVEUb8jTo6IniZS5Dn9kvGIRFgg2hWZaAkvogJZcKCMTRmnGNqcJp_9jupMcGQbNL8siS1Ifz4fnXNRSs1VXwF44tcZHzy0wP47hmhBUIPniYcMKlbTGlre1dfxSMhe904tXKvM0561PoplC3YqgUseCDY8ApsNOl7syiOQisB12Y; OGPC=19008104-1:19008862-2:19009193-2:19009353-1:873035776-1:; OTZ=4682697_48_48_123840_44_436320; OGP=-19009193:-19009353:; _gcl_au=1.1.529736880.1542033104; SID=swZLJh3gRc_fGHq2NrqgGwGNHxM6wduy9mAEpZuD3jGi1eEqThjQ_n9wRZjXAGjH2iAGVw.; APISID=1Q2Nkf9qRFQi4Dj7/AyzvBSHrL2CKhmIja; HSID=AxzsCIE3E8Id9Q9Ht; SAPISID=QHWSWIFzLO8-kfXy/At3vNrfL-RLGnY9GT; SSID=Alw4UteDDApBqtosx; _ga=GA1.1.880952510.1527575116")
		c.SetHeader(header)

		rawUrl := "https://www.cbr-xml-daily.ru/daily_json.js"
		err = c.Request(http.MethodGet, rawUrl, nil)
		if err != nil {
			return
		}

		var status uint
		status = c.GetStatus()
		if status != 200 {
			err = fmt.Errorf("get unexpected status on `%s`", rawUrl)

			return
		}

		var bs []byte
		bs, err = io.ReadAll(c.GetBody())
		if err != nil {
			return
		}

		var list List
		err = json.Unmarshal(bs, &list)
		if err != nil {
			return
		}

		for name, currency := range list.Currencies {
			currencies.Store(name, currency)
		}
	}

	from = strings.ToUpper(from)

	to = strings.ToUpper(to)

	fromIface, ok := currencies.Load(from)
	if !ok {
		err = fmt.Errorf("unexpected currency `from`: %s", from)

		return
	}

	fromCurrency := fromIface.(Currency)

	toIface, ok := currencies.Load(to)
	if !ok {
		err = fmt.Errorf("unexpected currency `to`: %s", from)

		return
	}

	toCurrency := toIface.(Currency)

	factor = (fromCurrency.Value / float64(fromCurrency.Nominal)) / (toCurrency.Value / float64(toCurrency.Nominal))

	return
}
