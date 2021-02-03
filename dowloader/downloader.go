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
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
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
		chromedp.WaitVisible(".welcome__agree"),
		chromedp.Click(`.welcome__agree`, chromedp.NodeVisible),
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible(".welcome__button--accept"),
		chromedp.Click(`.welcome__button--accept`, chromedp.NodeVisible),
		chromedp.Sleep(time.Second),
		chromedp.WaitVisible(".transfer__button"),
		chromedp.Click(`.transfer__button`, chromedp.NodeVisible),
		chromedp.Sleep(time.Minute),
	)

	return
}
