package xlsx

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/leprosus/golang-log"
	"strings"
)

type Parser struct {
	file *excelize.File
}

func NewParser(filePath string) (parser *Parser, err error) {
	parser = &Parser{}
	parser.file, err = excelize.OpenFile(filePath)

	return
}

func (parser *Parser) Parse(sheetName string) (data Data, headers Headers, err error) {
	if !parser.IsSheetExist(sheetName) {
		return
	}

	var rows [][]string
	rows, err = parser.file.GetRows(sheetName)
	if err != nil || len(rows) == 0 {
		return
	}

	var headersRowId int
	headersRowId, err = defineRowOfHeaders(rows)
	if err != nil {
		return
	}

	headers = extractHeader(rows[headersRowId])

	if headersRowId >= len(rows) {
		return
	}

	rows = rows[headersRowId+1:]

	var (
		colId, rowId int
		row          []string
		cellName     string
		value        string

		imageName string
		bs        []byte

		newRow = Row{}

		header Header
		ok     bool
	)

	for rowId, row = range rows {
		if len(row) > 0 &&
			strings.TrimSpace(row[0]) == "" {
			continue
		}

		newRow = Row{}

		for colId, value = range row {
			cellName, err = excelize.CoordinatesToCellName(colId+1, rowId+1)
			if err != nil {
				return
			}

			imageName, bs, err = parser.file.GetPicture(sheetName, cellName)
			if err != nil {
				return
			}

			header, ok = headers[colId]
			if !ok {
				continue
			}

			newRow[header] = Cell{
				value:     value,
				imageName: imageName,
				bs:        bs,
			}
		}

		data = append(data, newRow)
	}

	return
}

func defineRowOfHeaders(rows [][]string) (rowId int, err error) {
	const rowThreshold = 20

	var (
		row   []string
		value string
	)
	for rowId, row = range rows {
		if rowId+1 > rowThreshold {
			break
		}

		for _, value = range row {
			value = strings.ToLower(value)
			value = strings.Replace(value, "_", "-", -1)

			if value == "asin" {
				return
			}
		}
	}

	err = fmt.Errorf("there aren't expected headers")

	return
}

func extractHeader(row []string) (headers Headers) {
	headers = Headers{}

	for colId, value := range row {
		value = strings.TrimSpace(value)

		if len(value) == 0 {
			continue
		}

		value = strings.ToLower(value)
		value = strings.Replace(value, "_", "-", -1)
		value = strings.Replace(value, "/ ", "-", -1)
		value = strings.Replace(value, "/", "-", -1)

		// Replacement
		switch value {
		case "asin":
			headers[colId] = ASIN
		case "brand":
			headers[colId] = Brand
		case "cat", "category":
			headers[colId] = Category
		case "sub category", "subcat":
			headers[colId] = SubCategory
		case "color":
			headers[colId] = Color
		case "condition":
			headers[colId] = Condition
		case "qty", "menge", "quantity":
			headers[colId] = Quantity
		case "bezeichnung", "itemname":
			headers[colId] = Title
		case "article description", "description":
			headers[colId] = Description
		case "bilder", "picture", "image", "image link":
			headers[colId] = Image
		case "vk netto", "vk netto alt", "cost":
			headers[colId] = Price
		case "vk netto neu", "price net":
			headers[colId] = PriceNet
		case "retail price":
			headers[colId] = RetailPrice
		case "ean code", "ean":
			headers[colId] = EAN
		case "price-unit":
			headers[colId] = UnitPrice
		case "price-total":
			headers[colId] = TotalPrice
		case "gender":
			headers[colId] = Gender
		case "lager-id":
			headers[colId] = LagerId
		case "land":
			headers[colId] = Land
		case "size":
			headers[colId] = Size
		case "sku":
			headers[colId] = SKU
		case "units":
			headers[colId] = Units
		case "weight":
			headers[colId] = Weight
		default:
			headers[colId] = Header(value)

			log.WarnFmt("Get unexpected header: %s", value)
		}
	}

	return
}

func (parser *Parser) IsSheetExist(sheetName string) (ok bool) {
	sheetMap := parser.file.GetSheetMap()
	for _, sheet := range sheetMap {
		if sheet == sheetName {
			return true
		}
	}

	return false
}
