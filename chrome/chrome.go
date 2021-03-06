package chrome

import (
	"context"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"io/ioutil"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type Chrome struct {
	isHeadless bool

	withProxy bool
	proxyAddr string
	proxyUser string
	proxyPass string

	timeout time.Duration

	ctx    context.Context
	cancel context.CancelFunc

	path string

	headers map[string]interface{}

	acceptedCodes map[uint32]interface{}
}

func New() (c *Chrome) {
	return &Chrome{
		isHeadless: true,
		timeout:    time.Minute,
		headers: map[string]interface{}{
			"accept-language": "en-US,en;q=0.9,de;q=0.8,fr;q=0.7,it;q=0.6,es;q=0.5,nl;q=0.4,*;q=0.2",
		},
		acceptedCodes: map[uint32]interface{}{
			200: nil,
			301: nil,
			302: nil,
		},
	}
}

func (b *Chrome) Proxy(proxyAddr string) (err error) {
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

func (b *Chrome) UserAgent(userAgent string) {
	b.SetHeader("user-agent", userAgent)
}

func (b *Chrome) SetHeader(header, value string) {
	b.headers[header] = value
}

func (b *Chrome) AddAcceptedResponseCode(codes ...uint32) {
	for _, code := range codes {
		b.acceptedCodes[code] = nil
	}
}

func (b *Chrome) Timeout(timeout time.Duration) {
	b.timeout = timeout
}

func (b *Chrome) Headless(isHeadless bool) {
	b.isHeadless = isHeadless
}

func (b *Chrome) LastScreenshot(path string) (err error) {
	b.path, err = filepath.Abs(path)
	if err != nil {
		return
	}

	err = os.MkdirAll(b.path, 0755)
	if err != nil {
		return
	}

	return
}

func (b *Chrome) Cancel() {
	b.cancel()

	return
}

func (b *Chrome) Run(rawUrl string, actions []chromedp.Action) (err error) {
	opts := chromedp.DefaultExecAllocatorOptions[:]

	if b.withProxy {
		opts = append(opts, chromedp.ProxyServer(b.proxyAddr))
	}

	opts = append(opts, chromedp.Flag("headless", b.isHeadless),
		chromedp.Flag("incognito", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("ignore-certificate-errors", true),
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

				_ = fetch.ContinueRequest(ev.RequestID).Do(execCtx)
			case *network.EventResponseReceived:
				if ev.Type == network.ResourceTypeDocument {
					_, ok := b.acceptedCodes[uint32(ev.Response.Status)]
					if !ok {
						b.Cancel()
					}
				}
			}
		}()
	})

	if len(b.path) > 0 {
		var bs []byte
		defer func() {
			if len(bs) == 0 {
				return
			}

			file, e := ioutil.TempFile(b.path, "screenshot-*.png")
			if e != nil {
				err = errors.Wrap(err, e.Error())

				return
			}

			_, e = file.Write(bs)
			if e != nil {
				err = errors.Wrap(err, e.Error())

				return
			}
		}()

		actions = append([]chromedp.Action{chromedp.ActionFunc(func(ctx context.Context) (err error) {
			go func() {
				for {
					err = FullScreenshot(100, &bs).Do(ctx)
					if err != nil {
						return
					}

					time.Sleep(100 * time.Millisecond)
				}
			}()

			return
		})}, actions...)
	}

	actions = append([]chromedp.Action{
		fetch.Enable().WithHandleAuthRequests(true),
		network.Enable(),
		network.SetExtraHTTPHeaders(b.headers),
		chromedp.Navigate(rawUrl),
	}, actions...)

	err = chromedp.Run(b.ctx, actions...)

	return
}

func (b *Chrome) GetHtml(rawUrl string) (html string, err error) {
	err = b.Run(rawUrl, []chromedp.Action{
		chromedp.OuterHTML("html", &html),
	})

	return
}

func (b *Chrome) MakeFullScreenshot(rawUrl string, quality int64) (bs []byte, err error) {
	err = b.Run(rawUrl, FullScreenshot(quality, &bs))

	return
}

func FullScreenshot(quality int64, bs *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) (err error) {
			var contentSize *dom.Rect
			_, _, contentSize, err = page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return
			}

			*bs, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return
			}

			return
		}),
	}
}
