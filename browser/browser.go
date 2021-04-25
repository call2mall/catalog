package browser

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
	"sync"
	"time"
)

type Browser struct {
	isInit   bool
	once     *sync.Once
	launcher *launcher.Launcher
	browser  *rod.Browser
	tab      *Tab

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

func NewBrowser() (c *Browser) {
	return &Browser{
		tab:  &Tab{},
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
	if b.isInit {
		b.browser.MustClose()
		b.launcher.Cleanup()
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

		if b.slowMotionTime > 0 {
			b.browser = b.browser.SlowMotion(b.slowMotionTime)
		}

		b.browser = b.browser.MustConnect()
	})
	if err != nil {
		return
	}

	b.tab.launcher = b.launcher
	b.tab.browser = b.browser

	b.tab.page, b.tab.cancelFunc = stealth.MustPage(b.tab.browser).WithCancel()

	b.isInit = true

	err = cb(b.tab)
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
