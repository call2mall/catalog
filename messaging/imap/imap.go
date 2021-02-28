package imap

import (
	"fmt"
	"github.com/BrianLeishman/go-imap"
	"time"
)

type Client struct {
	dialer *imap.Dialer
	mark   string
	folder string
}

type Email struct {
	Flags       []string
	Received    time.Time
	Sent        time.Time
	Size        uint64
	Subject     string
	UID         uint32
	MessageID   string
	From        EmailAddr
	To          EmailAddr
	ReplyTo     EmailAddr
	CC          EmailAddr
	BCC         EmailAddr
	Text        string
	HTML        string
	Attachments []Attach
}

type EmailAddr map[string]string

type Attach struct {
	Name     string
	MimeType string
	Content  []byte
}

func NewClient(username, password, host string, port uint) (c *Client, err error) {
	c = &Client{}
	c.dialer, err = imap.New(username, password, host, int(port))
	c.folder = "INBOX"

	return
}

func (c *Client) SetSearchMark(mark string) {
	c.mark = mark
}

func (c *Client) SetFolder(folder string) {
	c.folder = folder
}

func (c *Client) Close() (err error) {
	return c.dialer.Close()
}

func conv(ext imap.EmailAddresses) (inn EmailAddr) {
	inn = EmailAddr{}

	for key, val := range ext {
		inn[key] = val
	}

	return
}

func (c *Client) GetAllEmails(lastUID int) (list []Email, err error) {
	err = c.dialer.SelectFolder(c.folder)
	if err != nil {
		return
	}

	var uids []int
	uids, err = c.dialer.GetUIDs(fmt.Sprintf("TEXT %s", c.mark))
	if err != nil {
		return
	}

	if lastUID > -1 {
		for i := 0; i < len(uids); i++ {
			if uids[i] <= lastUID {
				uids = append(uids[:i], uids[i+1:]...)
				i--
			}
		}
	}

	var emails map[int]*imap.Email
	emails, err = c.dialer.GetEmails(uids...)
	if err != nil {
		return
	}

	var attachList []Attach
	for _, one := range emails {
		attachList = []Attach{}

		for _, attach := range one.Attachments {
			attachList = append(attachList, Attach{
				Name:     attach.Name,
				MimeType: attach.MimeType,
				Content:  attach.Content,
			})
		}

		list = append(list, Email{
			Flags:       one.Flags,
			Received:    one.Received,
			Sent:        one.Sent,
			Size:        one.Size,
			Subject:     one.Subject,
			UID:         uint32(one.UID),
			MessageID:   one.MessageID,
			From:        conv(one.From),
			To:          conv(one.To),
			ReplyTo:     conv(one.ReplyTo),
			CC:          conv(one.CC),
			BCC:         conv(one.BCC),
			Text:        one.Text,
			HTML:        one.HTML,
			Attachments: attachList,
		})
	}

	return
}

func (c *Client) GetLatestEmail(lastUID int) (email Email, err error) {
	var list []Email
	list, err = c.GetAllEmails(lastUID)
	if err != nil {
		return
	}

	var maxUID uint32
	for _, one := range list {
		if one.UID > maxUID {
			maxUID = one.UID
			email = one
		}
	}

	return
}
