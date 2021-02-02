package parser

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/leprosus/golang-log"
	"regexp"
	"strings"
)

type Parser struct {
	file *excelize.File

	data    map[string]Data
	headers map[string]Headers
}

func NewParser(filePath string) (parser *Parser, err error) {
	parser = &Parser{
		data:    map[string]Data{},
		headers: map[string]Headers{},
	}
	parser.file, err = excelize.OpenFile(filePath)

	return
}

func (p *Parser) Parse(sheetName string) (data Data, headers Headers, err error) {
	if !p.IsSheetExist(sheetName) {
		return
	}

	_, ok := p.data[sheetName]
	if ok {
		data = p.data[sheetName]
		headers = p.headers[sheetName]

		return
	}

	var rawData [][]string
	rawData, err = p.GetRawData(sheetName)
	if err != nil {
		return
	}

	var headersRowId int
	headersRowId, err = p.defineRowOfHeaders(rawData)
	if err != nil {
		return
	}

	if headersRowId >= len(rawData) {
		return
	}

	headers = p.extractHeader(rawData[headersRowId])
	rawData = rawData[headersRowId+1:]

	var (
		colId           int
		cellName, value string

		imageName string
		bs        []byte

		newRow Row
		header Header
	)

	for rowId, row := range rawData {
		newRow = Row{}

		for colId, value = range row {
			cellName, err = excelize.CoordinatesToCellName(colId+1, rowId+1)
			if err != nil {
				return
			}

			imageName, bs, err = p.file.GetPicture(sheetName, cellName)
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

	p.data[sheetName] = data
	p.headers[sheetName] = headers

	return
}

func (p *Parser) GetRawData(sheetName string) (data [][]string, err error) {
	var rows [][]string
	rows, err = p.file.GetRows(sheetName)
	if err != nil || len(rows) == 0 {
		return
	}

	exceptRowsIds := findEmptyLinesIds(rows, 50)

	var (
		cleanData [][]string
		ok        bool
	)
	for rowId, row := range rows {
		_, ok = exceptRowsIds[rowId]
		if ok {
			continue
		}

		cleanData = append(cleanData, row)
	}

	exceptColsIds := findEmptyLinesIds(transposeData(cleanData), 10)

	var (
		colId int
		value string
		line  []string
	)
	for _, row := range cleanData {
		line = []string{}

		for colId, value = range row {
			_, ok = exceptColsIds[colId]
			if ok {
				continue
			}

			line = append(line, value)
		}

		if len(line) > 0 {
			data = append(data, line)
		}
	}

	return
}

func findEmptyLinesIds(data [][]string, threshold float32) (ids map[int]interface{}) {
	ids = map[int]interface{}{}

	var (
		total, nonempty uint
		value           string
	)

	for id, line := range data {
		total = 0
		nonempty = 0

		for _, value = range line {
			if strings.TrimSpace(value) != "" {
				nonempty++
			}

			total++
		}

		if nonempty == 0 || float32(nonempty)*100/float32(total) < threshold {
			ids[id] = nil
		}
	}

	return
}

func transposeData(data [][]string) (transposed [][]string) {
	var (
		rowLen, colLen int
		row            []string
	)

	rowLen = len(data)

	for _, row = range data {
		if len(row) > colLen {
			colLen = len(row)
		}
	}

	var (
		vector       []string
		rowId, colId int
		value        string
	)

	for rowId = 0; rowId < rowLen; rowId++ {
		row = data[rowId]

		for colId = 0; colId < colLen; colId++ {
			value = ""

			if len(row) > colId {
				value = row[colId]
			}

			vector = append(vector, value)
		}
	}

	var (
		id, offset int
		line       []string

		total = rowLen * colLen
	)

	for offset = 0; offset < colLen; offset++ {
		line = []string{}
		for id = 0; id < total; id += colLen {
			line = append(line, vector[id+offset])
		}

		transposed = append(transposed, line)
	}

	return
}

func (p *Parser) defineRowOfHeaders(rows [][]string) (rowId int, err error) {
	const rowsThreshold = 20

	var (
		row   []string
		value string
	)
	for rowId, row = range rows {
		if rowId+1 > rowsThreshold {
			break
		}

		for _, value = range row {
			value = clearHeaderField(value)

			if value == "lager id" || value == "ean" || value == "sku" || value == "asin" ||
				value == "title" || value == "description" {
				return
			}
		}
	}

	err = fmt.Errorf("there aren't expected headers")

	return
}

func (p *Parser) extractHeader(line []string) (headers Headers) {
	headers = Headers{}

	for id, value := range line {
		value = strings.TrimSpace(value)

		if len(value) == 0 {
			continue
		}

		value = clearHeaderField(value)

		// Replacement
		switch value {
		case "asin":
			headers[id] = ASIN
		case "brand":
			headers[id] = Brand
		case "cat", "category":
			headers[id] = Category
		case "sub category", "subcat":
			headers[id] = SubCategory
		case "color":
			headers[id] = Color
		case "condition":
			headers[id] = Condition
		case "qty", "menge", "quantity":
			headers[id] = Quantity
		case "bezeichnung", "itemname":
			headers[id] = Title
		case "article description", "description":
			headers[id] = Description
		case "bilder", "picture", "image", "image link":
			headers[id] = Image
		case "retail price":
			headers[id] = RetailPrice
		case "ean code", "ean":
			headers[id] = EAN
		case "vk netto", "vk netto alt", "cost", "avides price", "price unit", "vk netto neu", "price net", "preis netto":
			headers[id] = UnitPrice
		case "price total", "preis total", "avides total":
			headers[id] = TotalPrice
		case "gender":
			headers[id] = Gender
		case "lager id":
			headers[id] = LagerId
		case "land":
			headers[id] = Land
		case "size":
			headers[id] = Size
		case "sku":
			headers[id] = SKU
		case "units":
			headers[id] = Units
		case "weight":
			headers[id] = Weight
		case "fassung":
			headers[id] = Type
		default:
			headers[id] = Header(value)

			log.WarnFmt("Get unexpected header: %s", value)
		}
	}

	return
}

var cleanRegExp = regexp.MustCompile("[\\-_/\\s]+")

func clearHeaderField(value string) string {
	return cleanRegExp.ReplaceAllString(strings.ToLower(value), " ")
}

func (p *Parser) IsSheetExist(sheetName string) (ok bool) {
	sheetMap := p.file.GetSheetMap()
	for _, sheet := range sheetMap {
		if sheet == sheetName {
			return true
		}
	}

	return false
}
