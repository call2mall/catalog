package dowloader

import (
	"context"
	"fmt"
	"github.com/call2mall/catalog/proxy"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/leprosus/golang-log"
	"net/url"
	"time"
)

func DownloadFromWetransfer(rawUrl string, proxies *proxy.Proxies, archivePath string) (err error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
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
				log.ErrorFmt("Catch error on proxy auth: %v", err)
			}
		}()
	})

	err = chromedp.Run(ctx,
		fetch.Enable().WithHandleAuthRequests(true),
		page.SetDownloadBehavior(page.SetDownloadBehaviorBehaviorAllow).WithDownloadPath(archivePath),
		chromedp.Navigate(rawUrl),

		chromedp.WaitVisible(".logo"),
		chromedp.Sleep(time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				_ = chromedp.WaitVisible(".welcome__agree").Do(ctx)
				_ = chromedp.Click(`.welcome__agree`, chromedp.NodeVisible).Do(ctx)
			}()

			return
		}),
		chromedp.Sleep(time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				_ = chromedp.WaitVisible(".welcome__button--accept").Do(ctx)
				_ = chromedp.Click(`.welcome__button--accept`, chromedp.NodeVisible).Do(ctx)
			}()

			return
		}),
		chromedp.Sleep(time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				_ = chromedp.WaitVisible(".transfer__button").Do(ctx)
				_ = chromedp.Click(`.transfer__button`, chromedp.NodeVisible).Do(ctx)
			}()

			return
		}),
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible(".transfer__button"),
		chromedp.Click(`.transfer__button`, chromedp.NodeVisible),
		chromedp.Sleep(time.Minute),
	)

	return
}
