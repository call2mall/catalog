package grabber

import (
	. "github.com/call2mall/catalog/messaging/imap"
	"github.com/leprosus/golang-config"
)

func getIMAPInstance() (imap *Client, err error) {
	imap, err = NewClient(config.String("imap.user"), config.String("imap.pass"), config.String("imap.host"), uint(config.Int32("imap.port")))
	if err != nil {
		return
	}
	defer func() {
		_ = imap.Close()
	}()

	imap.SetFolder(config.String("imap.folder"))

	imap.SetSearchMark(config.String("imap.mark"))

	return
}
