package utils

import (
	"fmt"
	"testing"
)

func TestWalk(t *testing.T) {
	list, err := Walk("../docs", "xlsx")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(list)
}
