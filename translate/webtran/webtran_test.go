package webtran

import (
	"fmt"
	"testing"
)

func TestTranslate(t *testing.T) {
	tr := WebTran{}

	text, err := tr.Translate("This is me", "en", "ru", "http://emiles01:xVypbJnv@51.83.17.111:29842")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(text)
}
