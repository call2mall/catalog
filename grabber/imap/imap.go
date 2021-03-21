package imap

import (
	"github.com/call2mall/catalog/dao"
	. "github.com/call2mall/catalog/messaging/imap"
	"github.com/leprosus/golang-config"
	"regexp"
)

func LoadAllEmails() (list []Email, err error) {
	var imapClient *Client
	imapClient, err = getIMAPInstance()
	if err != nil {
		return
	}

	list, err = imapClient.GetAllEmails(-1)

	return
}

func LoadLatestEmail() (email Email, err error) {
	var imapClient *Client
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

var wetransferRegExp = regexp.MustCompile("https://we.tl/[a-zA-Z\\-\\d]+\\b")

func ExtractWetransferUrl(text string) (rawUrl string) {
	matches := wetransferRegExp.FindAllString(text, -1)

	if len(matches) == 0 {
		return
	}

	rawUrl = matches[0]

	return
}
