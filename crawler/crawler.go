package crawler

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Crawler struct {
	client  http.Client
	header  http.Header
	ctx     context.Context
	withCtx bool
	req     *http.Request
	res     *http.Response
}

func NewCrawler() (c *Crawler) {
	c = &Crawler{
		client: http.Client{},
	}

	return
}

func (c *Crawler) SetHeader(header http.Header) {
	c.header = header
}

func (c *Crawler) AddHeader(key, value string) {
	c.header.Add(key, value)
}

func (c *Crawler) SetUserAgent(userAgent string) {
	c.AddHeader("user-agent", userAgent)
}

func (c *Crawler) SetTimeout(timeout time.Duration) {
	c.client.Timeout = timeout
}

func (c *Crawler) SetProxy(proxyAddr string) (err error) {
	var proxyUrl *url.URL
	proxyUrl, err = url.Parse(proxyAddr)
	if err != nil {
		return
	}

	c.client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyUrl),
	}

	return
}

func (c *Crawler) SetContext(ctx context.Context) {
	c.ctx = ctx
	c.withCtx = true
}

func (c *Crawler) Request(method, rawUrl string, body io.Reader) (err error) {
	if c.withCtx {
		c.req, err = http.NewRequestWithContext(c.ctx, method, rawUrl, body)
	} else {
		c.req, err = http.NewRequest(method, rawUrl, body)
	}
	if err != nil {
		return
	}

	c.req.Header = c.header

	c.res, err = c.client.Do(c.req)
	if err != nil {
		return
	}

	return
}

func (c *Crawler) GetStatus() (status uint) {
	if c.res == nil {
		return
	}

	status = uint(c.res.StatusCode)

	return
}

func (c *Crawler) GetHeader() (header http.Header) {
	header = c.res.Header

	return
}

func (c *Crawler) GetBody() (body io.ReadCloser) {
	body = c.res.Body

	return
}
