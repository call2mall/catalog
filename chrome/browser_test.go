package chrome

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestBrowser(t *testing.T) {
	rawUrl := "https://www.amazon.de/review/product/B07Q3S8BKF"
	b := NewBrowser()
	b.Headless(false)
	err := b.Proxy("http://emiles01:xVypbJnv@51.89.130.34:29842")
	if err != nil {
		t.Error(err)
	}

	var html string
	html, err = b.GetHtml(rawUrl)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(html)
}

func TestFullScreenshot(t *testing.T) {
	rawUrl := "https://www.amazon.de/review/product/B07Q3S8BKF"
	b := NewBrowser()
	b.Headless(false)
	err := b.Proxy("http://emiles01:xVypbJnv@51.89.130.34:29842")
	if err != nil {
		t.Error(err)
	}

	defer func() {
		_ = os.RemoveAll("screenshot.png")
	}()

	var bs []byte
	bs, err = b.MakeFullScreenshot(rawUrl, 100)
	if err != nil {
		t.Error(err)
	}

	err = ioutil.WriteFile("screenshot.png", bs, 0644)
	if err != nil {
		t.Error(err)
	}
}
