package currency

import (
	"testing"
)

func TestGetFactor(t *testing.T) {
	factor, err := GetFactor("gbp", "eur")
	if err != nil {
		t.Fatal(err.Error())
	}

	if factor == 0.0 {
		t.Fatal("Can't convert GBP -> EUR")
	}
}
