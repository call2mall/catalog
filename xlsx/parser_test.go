package xlsx

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	parser, err := NewParser("../docs/wetransfer-49cd36/Amazon Beauty Offer.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(parser.IsSheetExist("Packinglist"))

	var (
		data    Data
		headers Headers
	)
	data, headers, err = parser.Parse("Packinglist")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(headers)

	row := data[0]
	fmt.Println(row)
}
