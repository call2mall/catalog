package browser

import (
	"fmt"
	"testing"
	"time"
)

func TestBrowser(t *testing.T) {
	source := "Как же всё заебало"

	b := NewBrowser()
	b.Headless(false)
	b.WithDevTools(true)
	b.WithTrace(true)
	b.WithSlowMotion(time.Second)

	err := b.Run(func(t *Tab) (err error) {
		err = t.Page().Navigate("https://translate.google.com")
		if err != nil {
			return
		}

		//btn := t.Page().MustElement(`[aria-label="Agree to the use of cookies and other data for the purposes described"]`)
		//btn.MustClick()

		el := t.Page().MustElement(`textarea[aria-label="Source text"]`)

		wait := t.Page().MustWaitRequestIdle("https://accounts.google.com")
		el.MustInput(source)
		wait()

		result := t.Page().MustElement("[role=region] span[lang]").MustText()

		fmt.Println(">", result)

		t.Page().MustClose()

		return
	})
	if err != nil {
		t.Fatal(err)
	}

	b.Close()
}
