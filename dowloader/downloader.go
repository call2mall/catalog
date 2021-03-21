package dowloader

import (
	"context"
	"github.com/call2mall/catalog/browser"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/pkg/errors"
	"time"
)

var TransferExpired = errors.New("transfer expired")

func DownloadFromWetransfer(rawUrl, downloadDir string, b *browser.Browser) (err error) {
	var isExpired bool

	err = b.Run(rawUrl, []chromedp.Action{
		fetch.Enable().WithHandleAuthRequests(true),
		page.SetDownloadBehavior(page.SetDownloadBehaviorBehaviorAllow).WithDownloadPath(downloadDir),
		chromedp.Navigate(rawUrl),

		chromedp.WaitVisible(".root-node", chromedp.ByQuery),
		chromedp.Sleep(time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				err = chromedp.WaitVisible(".welcome__agree", chromedp.ByQuery).Do(ctx)
				if err != nil {
					return
				}

				err = chromedp.Click(".welcome__agree", chromedp.ByQuery).Do(ctx)
				if err != nil {
					return
				}
			}()

			return
		}),
		chromedp.Sleep(time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				err = chromedp.WaitVisible(".welcome__button--accept", chromedp.ByQuery).Do(ctx)
				if err != nil {
					return
				}

				err = chromedp.Click(".welcome__button--accept", chromedp.ByQuery).Do(ctx)
				if err != nil {
					return
				}
			}()

			return
		}),
		chromedp.Sleep(time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				err = chromedp.WaitVisible(".downloader__expired", chromedp.ByQuery).Do(ctx)
				if err != nil {
					return
				}

				isExpired = true

				b.Cancel()
			}()

			return
		}),
		chromedp.Sleep(time.Second),
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				for {

					err = chromedp.WaitVisible(".transfer__button", chromedp.ByQuery).Do(ctx)
					if err != nil {
						return
					}

					err = chromedp.Click(".transfer__button", chromedp.ByQuery).Do(ctx)
					if err != nil {
						return
					}
				}
			}()

			return
		}),
		chromedp.Sleep(time.Minute),
	})
	if isExpired {
		err = TransferExpired

		return
	}

	if err != nil {
		return
	}

	return
}
