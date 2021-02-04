package parser

import (
	"fmt"
	"os"
	"testing"
)

func TestParser(t *testing.T) {
	file, err := os.Open("../docs/210112 Osram LED Leuchtmittel.xlsx")
	if err != nil {
		t.Fatal(err)
	}

	var parser *Parser
	parser, err = NewParser(file)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(parser.IsSheetExist("Packinglist"))

	var (
		data    Data
		headers Headers
	)
	data, headers, err = parser.Parse("Overview")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(headers)

	if len(data) > 0 {
		fmt.Println(data[0])
	}
}
