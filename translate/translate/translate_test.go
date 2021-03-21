package translate

import (
	"fmt"
	"testing"
)

func TestTranslate(t *testing.T) {
	tr := Translate{}

	text, err := tr.Translate("This is me", "en", "ru", "http://emiles01:xVypbJnv@51.83.17.111:29842")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(text)
}

func TestDetectLang(t *testing.T) {
	tr := Translate{}

	lang, err := tr.DetectLang("This is me", "http://emiles01:xVypbJnv@51.83.17.111:29842")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(lang)
}
