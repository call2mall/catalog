package translate

import (
	"testing"
)

func TestDetectL8n(t *testing.T) {
	if DetectL8n("Eono Essentials男子サッカージャージ（サイズ8年）") != JapaneseL8n {
		t.Error("Can't detect necessary localisation")
	}
}
