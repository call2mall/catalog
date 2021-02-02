package extractor

import (
	"fmt"
	"github.com/call2mall/catalog/dao"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractor(t *testing.T) {
	matches, err := filepath.Glob("../docs/*.xlsx")
	if err != nil {
		t.Fatal(err)
	}

	var e *Extractor

	for _, path := range matches {
		if strings.Contains(path, "EICHHORN") ||
			strings.Contains(path, "Notebooks Lenovo New") ||
			strings.Contains(path, "White Goods Samsung") ||
			strings.Contains(path, "ASOS") ||
			strings.Contains(path, "TVs Thomson refurbished") ||
			strings.Contains(path, "Fashion Autumn Winter Offer") ||
			strings.Contains(path, "Mavi Jeans Fashion Offer") ||
			strings.Contains(path, "NA-KD Fashion Spring Summer") ||
			strings.Contains(path, "Shoes Brand Mix") ||
			strings.Contains(path, "Soccer Wear Offer") ||
			strings.Contains(path, "TomTailor Autumn Winter Fashion Offer") {
			continue
		}

		fmt.Println("Extracting of", path)

		e, err = NewExtractor(path)
		if err != nil {
			t.Fatal(err)
		}

		var list dao.SKUList
		list, err = e.Extract()
		if err != nil {
			t.Fatal(err)
		}

		if len(list) == 0 {
			fmt.Println("No one item has been found", path)
		}
	}
}
