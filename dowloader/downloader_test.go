package dowloader

import (
	"github.com/call2mall/catalog/chrome"
	"testing"
	"time"
)

func TestDownloadFromWetransfer(t *testing.T) {
	b := chrome.NewBrowser()
	b.Timeout(2 * time.Minute)
	b.Headless(false)

	err := DownloadFromWetransfer("https://we.tl/t-oIVZCqVB0v", ".", b)
	if err != nil {
		t.Fatal(err.Error())
	}
}
