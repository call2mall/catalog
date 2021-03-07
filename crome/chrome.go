package crome

import (
	"context"
	"fmt"
	"github.com/call2mall/catalog/proxy"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/leprosus/golang-log"
	"net/url"
	"time"
)

func Run(rawUrl string, actions []chromedp.Action, proxies *proxy.Proxies) (err error) {
	proxyAddr, ok := proxies.Next()
	if !ok {
		err = fmt.Errorf("can't get next proxy address")

		return
	}

	var proxyData *url.URL
	proxyData, err = url.Parse(proxyAddr)
	if err != nil {
		return
	}

	userInfo := proxyData.User
	proxyData.User = nil

	addr := proxyData.String()

	username := userInfo.Username()
	password, _ := userInfo.Password()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer(addr),
		chromedp.Flag("headless", true),
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

	ctx, cancel = chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	defer func() {
		_ = chromedp.Cancel(ctx)
	}()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		go func() {
			var err error

			switch ev := ev.(type) {
			case *fetch.EventAuthRequired:
				execCtx := cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Target)

				res := &fetch.AuthChallengeResponse{
					Response: fetch.AuthChallengeResponseResponseProvideCredentials,
					Username: username,
					Password: password,
				}

				err = fetch.ContinueWithAuth(ev.RequestID, res).Do(execCtx)
			case *fetch.EventRequestPaused:
				execCtx := cdp.WithExecutor(ctx, chromedp.FromContext(ctx).Target)

				err = fetch.ContinueRequest(ev.RequestID).Do(execCtx)
			}

			if err != nil {
				log.ErrorFmt("Catch error on proxy auth of `%s`: %v", proxyAddr, err)
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

	err = chromedp.Run(ctx, actions...)

	return
}

func GetHtml(rawUrl string, proxies *proxy.Proxies) (html string, err error) {
	err = Run(rawUrl, []chromedp.Action{
		chromedp.OuterHTML("html", &html),
	}, proxies)

	return
}
