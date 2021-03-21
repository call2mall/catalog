package browser

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/leprosus/golang-log"
	"net/url"
	"time"
)

type Browser struct {
	isHeadless bool

	withProxy bool
	proxyAddr string
	proxyUser string
	proxyPass string

	timeout time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

func NewBrowser() (c *Browser) {
	return &Browser{
		isHeadless: true,
		timeout:    time.Minute,
	}
}

func (b *Browser) Proxy(proxyAddr string) (err error) {
	if b.withProxy {
		return
	}

	var proxyData *url.URL
	proxyData, err = url.Parse(proxyAddr)
	if err != nil {
		return
	}

	userInfo := proxyData.User
	proxyData.User = nil

	b.withProxy = true
	b.proxyAddr = proxyData.String()
	b.proxyUser = userInfo.Username()
	b.proxyPass, _ = userInfo.Password()

	return
}

func (b *Browser) Timeout(timeout time.Duration) {
	b.timeout = timeout
}

func (b *Browser) Headless(isHeadless bool) {
	b.isHeadless = isHeadless
}

func (b *Browser) Cancel() {
	b.cancel()

	return
}

func (b *Browser) Run(rawUrl string, actions []chromedp.Action) (err error) {
	opts := chromedp.DefaultExecAllocatorOptions[:]

	if b.withProxy {
		opts = append(opts, chromedp.ProxyServer(b.proxyAddr))
	}

	opts = append(opts, chromedp.Flag("headless", b.isHeadless),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-gpu-shader-disk-cache", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("aggressive-cache-discard", true),
		chromedp.Flag("disable-cache", true),
		chromedp.Flag("disable-application-cache", true),
		chromedp.Flag("disable-offline-load-stale-cache", true),
		chromedp.Flag("disk-cache-dir", "/dev/null"),
		chromedp.Flag("media-cache-dir", "/dev/null"),
		chromedp.Flag("disk-cache-size", "1"),
		chromedp.Flag("media-cache-size", "1"),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-accelerated-2d-canvas", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-profile", true),
	)

	b.ctx, b.cancel = context.WithTimeout(context.Background(), b.timeout)
	defer b.cancel()

	b.ctx, b.cancel = chromedp.NewExecAllocator(b.ctx, opts...)
	defer b.cancel()

	b.ctx, b.cancel = chromedp.NewContext(b.ctx)
	defer b.cancel()
	defer func() {
		_ = chromedp.Cancel(b.ctx)
	}()

	chromedp.ListenTarget(b.ctx, func(ev interface{}) {
		go func() {
			var err error

			switch ev := ev.(type) {
			case *fetch.EventAuthRequired:
				if b.withProxy {
					execCtx := cdp.WithExecutor(b.ctx, chromedp.FromContext(b.ctx).Target)

					res := &fetch.AuthChallengeResponse{
						Response: fetch.AuthChallengeResponseResponseProvideCredentials,
						Username: b.proxyUser,
						Password: b.proxyPass,
					}

					err = fetch.ContinueWithAuth(ev.RequestID, res).Do(execCtx)
					if err != nil {
						log.ErrorFmt("Catch error on proxy auth of `%s`: %v", b.proxyAddr, err)
					}
				}
			case *fetch.EventRequestPaused:
				execCtx := cdp.WithExecutor(b.ctx, chromedp.FromContext(b.ctx).Target)

				err = fetch.ContinueRequest(ev.RequestID).Do(execCtx)
			}

			if err != nil {
				log.Error(err.Error())
			}
		}()
	})

	var headers = map[string]interface{}{
		"accept-language": "en-US,en;q=0.9,de;q=0.8,fr;q=0.7,it;q=0.6,es;q=0.5,nl;q=0.4,*;q=0.2",
	}

	actions = append([]chromedp.Action{
		fetch.Enable().WithHandleAuthRequests(true),
		network.Enable(),
		network.SetExtraHTTPHeaders(headers),
		chromedp.Navigate(rawUrl),
	}, actions...)

	err = chromedp.Run(b.ctx, actions...)

	return
}

func (b *Browser) GetHtml(rawUrl string) (html string, err error) {
	err = b.Run(rawUrl, []chromedp.Action{
		chromedp.OuterHTML("html", &html),
	})

	return
}
