package grabber

import (
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/dowloader"
	"github.com/call2mall/catalog/extractor"
	"github.com/call2mall/catalog/messaging/imap"
	"github.com/call2mall/catalog/proxy"
	"github.com/call2mall/catalog/utils"
	"github.com/leprosus/golang-config"
	"github.com/leprosus/golang-log"
	"io"
	"net/http"
	"os"
	"regexp"
)

var (
	wetransferRegExp = regexp.MustCompile("https://we.tl/[a-zA-Z\\-\\d]+\\b")

	header = http.Header{}
)

func init() {
	header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
}

func GrabUnits() (err error) {
	dstPath := config.Path("grabber.path")
	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		return
	}

	var email imap.Email
	email, err = loadLatestEmail()
	if err != nil {
		return
	}

	if email.UID == 0 {
		return
	}

	srcUrl := extractWetransferUrl(email.Text)
	if len(srcUrl) == 0 {
		return
	}

	var ok bool
	ok, err = dao.IsProcessedWetransferUrl(srcUrl)
	if err != nil || ok {
		return
	}

	var proxies *proxy.Proxies
	proxies, err = proxy.GetInstance()
	if err != nil {
		return
	}

	err = dowloader.DownloadFromWetransfer(srcUrl, proxies, dstPath)
	if err != nil {
		return
	}

	err = dao.InsertWetransferUrl(email.UID, srcUrl)
	if err != nil {
		return
	}

	var unitList dao.UnitList
	unitList, err = extractPrices(dstPath)
	if err != nil || len(unitList) == 0 {
		return
	}

	asinList := unitList.ExtractASINList()

	err = asinList.Store()
	if err != nil {
		return
	}

	err = asinList.PushToSearchQueue()
	if err != nil {
		return
	}

	err = unitList.Store()
	if err != nil {
		return
	}

	log.DebugFmt("New %d units were stored successful", len(unitList))

	return
}

func loadAllEmails() (list []imap.Email, err error) {
	var imapClient *imap.Client
	imapClient, err = getIMAPInstance()
	if err != nil {
		return
	}

	list, err = imapClient.GetAllEmails(-1)

	return
}

func loadLatestEmail() (email imap.Email, err error) {
	var imapClient *imap.Client
	imapClient, err = getIMAPInstance()
	if err != nil {
		return
	}

	var lastUID uint32
	lastUID, err = dao.GetLastEmailUID()
	if err != nil {
		return
	}

	email, err = imapClient.GetLatestEmail(int(lastUID))

	return
}

func extractWetransferUrl(text string) (rawUrl string) {
	matches := wetransferRegExp.FindAllString(text, -1)

	if len(matches) == 0 {
		return
	}

	rawUrl = matches[0]

	return
}

func extractPrices(dirPath string) (list dao.UnitList, err error) {
	var filesPaths []string
	filesPaths, err = utils.Walk(dirPath, "zip")
	if err != nil {
		return
	}

	var (
		ex      *extractor.Extractor
		skuPart dao.UnitList
	)
	for _, filePath := range filesPaths {
		err = utils.Unzip(filePath, func(reader io.ReadCloser, fileName string) (err error) {
			ex, err = extractor.NewExtractor(reader)
			if err != nil {
				return
			}

			skuPart, err = ex.Extract()
			if err != nil {
				log.WarnFmt("Can't extract data from `%s` file `%s`: %s", filePath, fileName, err.Error())

				err = nil

				return
			}

			if len(skuPart) == 0 {
				return
			}

			list = append(list, skuPart...)

			return
		})
		if err != nil {
			return
		}

		err = os.Remove(filePath)
		if err != nil {
			return
		}
	}

	return
}
