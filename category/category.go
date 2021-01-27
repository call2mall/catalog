package category

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type Category struct {
	Id        uint32
	Name      string
	IsDefined bool

	Features
}

type Features struct {
	IsSale      bool
	IsExpensive bool
	Condition   Condition
}

func ExtractCategory(filePath string) (category Category, err error) {
	fileName := filepath.Base(filePath)

	pos := strings.Index(fileName, " ")
	l := len(fileName)
	if pos < 0 {
		err = fmt.Errorf("can't define category by file name of `%s`", filePath)

		return
	}

	fileName = fileName[pos+1 : l-5]
	fileName = strings.Replace(fileName, "_", " ", -1)

	name := strings.ToLower(fileName)

	if strings.Contains(name, "angebot") {
		category.IsSale = true

		name = strings.Replace(name, "angebot", "", 1)
	}

	if strings.Contains(name, "defekt") {
		category.Condition = Defective

		name = strings.Replace(name, "defekt", "", 1)
	} else if strings.Contains(name, "a-ware") {
		category.Condition = AWare

		name = strings.Replace(name, "a-ware", "", 1)
	} else if strings.Contains(name, "ungeprüfte retourware") {
		category.Condition = Unchecked

		name = strings.Replace(name, "ungeprüfte retourware", "", 1)
	} else if strings.Contains(name, "retourware") {
		category.Condition = Returned

		name = strings.Replace(name, "retourware", "", 1)
	} else if strings.Contains(name, "refurbished") {
		category.Condition = AWare

		name = strings.Replace(name, "refurbished", "", 1)
	}

	if strings.Contains(name, "teuer") ||
		strings.Contains(name, "luxus") ||
		strings.Contains(name, "luxury") {
		category.IsExpensive = true

		name = strings.Replace(name, "teuer", "", 1)
		name = strings.Replace(name, "luxus", "", 1)
	}

	if strings.Contains(name, "%") {
		category.IsSale = true

		name = regexp.MustCompile(`\s-?\d+%`).ReplaceAllString(name, "")
	}

	category.IsDefined = true

	if strings.Contains(name, "notebook") {
		category.Name = "Notebook"
	} else if strings.Contains(name, "tv") {
		category.Name = "TV"
	} else if strings.Contains(name, "luxus uhren") ||
		strings.Contains(name, "watches jewellery") {
		category.Name = "Luxury watches"
	} else if strings.Contains(name, "ecovacs") {
		category.Name = "Domestic tech"
	} else if strings.Contains(name, "handy") {
		category.Name = "Smartphone"
	} else if strings.Contains(name, "pc") {
		category.Name = "PC"
	} else if strings.Contains(name, "schmuck") {
		category.Name = "Jewelry"
	} else if strings.Contains(name, "smartwatches") {
		category.Name = "Smartwatch"
	} else if strings.Contains(name, "tablets") {
		category.Name = "Tablet"
	} else if strings.Contains(name, "uhren") {
		category.Name = "Watch"
	} else if strings.Contains(name, "elektro roh") {
		category.Name = "Electrical"
	} else if strings.Contains(name, "elektronik") {
		category.Name = "Electronics"
	} else if strings.Contains(name, "toys") {
		category.Name = "Toy"
	} else if strings.Contains(name, "drogerie") {
		category.Name = "Medicine"
	} else if strings.Contains(name, "haushalt") {
		category.Name = "Domestic goods"
	} else if strings.Contains(name, "weisse") {
		category.Name = "Domestic tech"
	} else if strings.Contains(name, "spirituosen") {
		category.Name = "Alcohol"
	} else if strings.Contains(name, "britax") {
		category.Name = "Baby goods"
	} else if strings.Contains(name, "kinderwagen") {
		category.Name = "Pram"
	} else {
		category.Name = "Others"
		category.IsDefined = false
	}

	return
}
