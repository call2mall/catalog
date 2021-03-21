package grabber

import (
	"fmt"
	"github.com/call2mall/catalog/browser"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/dowloader"
	"github.com/call2mall/catalog/grabber/extractor"
	. "github.com/call2mall/catalog/grabber/imap"
	"github.com/call2mall/catalog/messaging/imap"
	"github.com/leprosus/golang-config"
	"github.com/leprosus/golang-log"
	"os"
	"time"
)

func GrabUnits() (err error) {
	dstPath := config.Path("grabber.path")
	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		return
	}

	var email imap.Email
	email, err = LoadLatestEmail()
	if err != nil {
		return
	}

	if email.UID == 0 {
		return
	}

	srcUrl := ExtractWetransferUrl(email.Text)
	if len(srcUrl) == 0 {
		return
	}

	var ok bool
	ok, err = dao.IsProcessedWetransferUrl(srcUrl)
	if err != nil || ok {
		return
	}

	defer func() {
		if err == nil {
			err = dao.InsertWetransferUrl(email.UID, srcUrl)
			if err != nil {
				return
			}
		}
	}()

	b := browser.NewBrowser()
	b.Timeout(2 * time.Minute)

	err = dowloader.DownloadFromWetransfer(srcUrl, dstPath, b)
	if err != nil {
		return
	}

	var unitList dao.UnitList
	unitList, err = extractor.ExtractData(dstPath)
	fmt.Println(len(unitList))
	if err != nil || len(unitList) == 0 {
		return
	}

	asinList := unitList.ExtractASINList()

	err = asinList.Store()
	if err != nil {
		return
	}

	err = unitList.Store()
	if err != nil {
		return
	}

	err = asinList.PushToSearchQueue()
	if err != nil {
		return
	}

	log.DebugFmt("New %d units were stored successful", len(unitList))

	return
}
