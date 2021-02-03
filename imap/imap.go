package main

import (
	"fmt"
	"github.com/BrianLeishman/go-imap"
	"time"
)

type Client struct {
	dialer *imap.Dialer
	mark   string
}

type Email struct {
	Flags       []string
	Received    time.Time
	Sent        time.Time
	Size        uint64
	Subject     string
	UID         int
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

	return
}

func (c *Client) SetSearchMark(mark string) {
	c.mark = mark
}

func (c *Client) Close() (err error) {
	return c.dialer.Close()
}

func (c *Client) GetLatestEmail(lastUID int) (email Email, err error) {
	err = c.dialer.SelectFolder("INBOX")
	if err != nil {
		return
	}

	var uids []int
	uids, err = c.dialer.GetUIDs(fmt.Sprintf("TEXT %s", c.mark))
	if err != nil {
		return
	}

	var maxUID int
	for _, id := range uids {
		if maxUID < id {
			maxUID = id
		}
	}

	if maxUID <= lastUID {
		return
	}

	var emails map[int]*imap.Email
	emails, err = c.dialer.GetEmails(maxUID)
	if err != nil {
		return
	}

	latest := emails[maxUID]

	conv := func(ext imap.EmailAddresses) (inn EmailAddr) {
		inn = EmailAddr{}

		for key, val := range ext {
			inn[key] = val
		}

		return
	}

	var attachList []Attach
	for _, attach := range latest.Attachments {
		attachList = append(attachList, Attach{
			Name:     attach.Name,
			MimeType: attach.MimeType,
			Content:  attach.Content,
		})
	}

	email = Email{
		Flags:       latest.Flags,
		Received:    latest.Received,
		Sent:        latest.Sent,
		Size:        latest.Size,
		Subject:     latest.Subject,
		UID:         latest.UID,
		MessageID:   latest.MessageID,
		From:        conv(latest.From),
		To:          conv(latest.To),
		ReplyTo:     conv(latest.ReplyTo),
		CC:          conv(latest.CC),
		BCC:         conv(latest.BCC),
		Text:        latest.Text,
		HTML:        latest.HTML,
		Attachments: attachList,
	}

	return
}
