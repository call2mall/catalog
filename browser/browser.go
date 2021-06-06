package browser

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"sync"
	"sync/atomic"
	"time"
)

type Browser struct {
	isInit, isClosed uint32
	once             *sync.Once
	launcher         *launcher.Launcher
	browser          *rod.Browser

	isHeadless     bool
	withDevTools   bool
	withTrace      bool
	slowMotionTime time.Duration

	header map[string]string
}

type Tab struct {
	launcher *launcher.Launcher
	browser  *rod.Browser
	page     *rod.Page

	cancelFunc func()
}

func New() (c *Browser) {
	return &Browser{
		once: &sync.Once{},

		isHeadless: true,
		header:     map[string]string{},
	}
}

func (b *Browser) Headless(isHeadless bool) {
	b.isHeadless = isHeadless
}

func (b *Browser) WithDevTools(withDevTools bool) {
	b.withDevTools = withDevTools
}

func (b *Browser) WithTrace(withTrace bool) {
	b.withTrace = withTrace
}

func (b *Browser) WithSlowMotion(duration time.Duration) {
	b.slowMotionTime = duration
}

func (b *Browser) Close() {
	if atomic.LoadUint32(&b.isInit) == 1 && atomic.LoadUint32(&b.isClosed) == 0 {
		b.browser.MustClose()
		b.launcher.Cleanup()

		atomic.StoreUint32(&b.isClosed, 1)
	}
}

func (b *Browser) Run(cb func(tab *Tab) (err error)) (err error) {
	b.once.Do(func() {
		b.launcher = launcher.New().
			Headless(b.isHeadless).
			Devtools(b.withDevTools)

		var controlUrl string
		controlUrl, err = b.launcher.Launch()
		if err != nil {
			return
		}

		b.browser = rod.New().
			ControlURL(controlUrl).
			Trace(b.withTrace)

		b.browser, err = b.browser.Incognito()
		if err != nil {
			return
		}

		if b.slowMotionTime > 0 {
			b.browser = b.browser.SlowMotion(b.slowMotionTime)
		}

		b.browser = b.browser.MustConnect()
	})
	if err != nil {
		return
	}

	tab := &Tab{
		launcher: b.launcher,
		browser:  b.browser,
	}

	tab.page, err = stealth.Page(tab.browser)
	if err != nil {
		return
	}

	tab.page, tab.cancelFunc = tab.page.WithCancel()

	atomic.StoreUint32(&b.isInit, 1)

	err = cb(tab)
	if err != nil {
		return
	}

	return
}

func (t *Tab) Launcher() (launcher *launcher.Launcher) {
	return t.launcher
}

func (t *Tab) Browser() (browser *rod.Browser) {
	return t.browser
}

func (t *Tab) Page() (page *rod.Page) {
	return t.page
}

func (t *Tab) SetHeader(key, value string) (err error) {
	_, err = t.page.SetExtraHeaders([]string{key, value})

	return
}

func (t *Tab) Close() {
	t.cancelFunc()
	t.Close()
}
