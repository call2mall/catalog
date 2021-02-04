package extractor

import (
	"fmt"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/parser"
	"github.com/leprosus/golang-log"
	"os"
)

type Extractor struct {
	filePath string
	parser   *parser.Parser
}

func NewExtractor(filePath string) (e *Extractor, err error) {
	e = &Extractor{}

	var file *os.File
	file, err = os.Open(filePath)
	if err != nil {
		return
	}

	e.filePath = filePath
	e.parser, err = parser.NewParser(file)
	if err != nil {
		return
	}

	return
}

func (e *Extractor) Extract() (list dao.SKUList, err error) {
	var priceType PriceType
	priceType, err = e.definePriceType()
	if err != nil {
		return
	}

	switch priceType {
	case PriceList:
		list, err = e.extractPriceList()
		if err != nil {
			return
		}
	case Collection:
		list, err = e.extractCollection()
		if err != nil {
			return
		}
	}

	return
}

type PriceType string

const (
	PriceList  PriceType = "price-list"
	Collection PriceType = "collection"
)

func (e *Extractor) definePriceType() (priceType PriceType, err error) {
	if !e.parser.IsSheetExist("Packinglist") {
		err = fmt.Errorf("document doesn't contain Packinglist sheet")

		return
	}

	var headers parser.Headers
	_, headers, err = e.parser.Parse("Packinglist")
	if err != nil {
		return
	}

	if headers.Contain(parser.LagerId, parser.EAN, parser.UnitPrice) {
		priceType = PriceList

		return
	}

	if e.parser.IsSheetExist("Overview") {
		priceType = Collection

		return
	}

	err = fmt.Errorf("unexpected type of document: %s", e.filePath)

	return
}

func (e *Extractor) extractPriceList() (list dao.SKUList, err error) {
	var data parser.Data
	data, _, err = e.parser.Parse("Packinglist")
	if err != nil {
		return
	}

	var (
		sku  dao.SKU
		cell parser.Cell
		ok   bool
	)
	for _, row := range data {
		sku = dao.SKU{}

		cell, ok = row[parser.ASIN]
		if !ok || cell.String() == "" {
			continue
		}
		sku.ASIN = cell.String()

		cell, ok = row[parser.LagerId]
		if ok {
			sku.WarehouseId = cell.String()
		}

		cell, ok = row[parser.SKU]
		if ok {
			sku.AvidesSKU = cell.String()
		}

		cell, ok = row[parser.EAN]
		if ok {
			sku.EAN = cell.String()
		}

		cell, ok = row[parser.Title]
		if ok {
			sku.Title = cell.String()
		}

		cell, ok = row[parser.Condition]
		if ok {
			sku.Condition = dao.ConvCondition(cell.String())
		}

		sku.Quantity = 1
		cell, ok = row[parser.Quantity]
		if ok {
			sku.Quantity, err = cell.UInt64()
			if err != nil {
				return
			}
		}

		cell, ok = row[parser.UnitPrice]
		if !ok {
			continue
		}
		sku.UnitCostInCent, err = cell.PriceInCent()
		if err != nil {
			return
		}

		if sku.Quantity > 1 {
			sku.UnitCostInCent /= sku.Quantity
		}

		list = append(list, sku)
	}

	return
}

func (e *Extractor) extractCollection() (list dao.SKUList, err error) {
	var (
		data    parser.Data
		headers parser.Headers

		intersect parser.Header
		ok        bool
	)

	data, headers, err = e.parser.Parse("Overview")
	if err != nil {
		return
	}

	intersect, ok = getIntersectedHeader(headers)
	if !ok {
		err = fmt.Errorf("there aren't header to intersect Overview and Packinglist sheets")

		return
	}

	var (
		cell  parser.Cell
		price uint64

		unitPriceByIntersect = map[string]uint64{}
		intersectValue       string
	)

	for _, row := range data {
		cell, ok = row[intersect]
		if !ok {
			err = fmt.Errorf("there aren't `%s` on Overview sheet", intersect)

			return
		}
		intersectValue = cell.String()

		cell, ok = row[parser.UnitPrice]
		if !ok {
			err = fmt.Errorf("there aren't `%s` on Overview sheet", parser.UnitPrice)

			return
		}

		price, err = cell.PriceInCent()
		if err != nil {
			return
		}

		unitPriceByIntersect[intersectValue] = price
	}

	data, _, err = e.parser.Parse("Packinglist")
	if err != nil {
		return
	}

	var sku dao.SKU
	for _, row := range data {
		sku = dao.SKU{}

		cell, ok = row[parser.ASIN]
		if !ok || cell.String() == "" {
			continue
		}
		sku.ASIN = cell.String()

		cell, ok = row[parser.LagerId]
		if ok {
			sku.WarehouseId = cell.String()
		}

		cell, ok = row[parser.SKU]
		if ok {
			sku.AvidesSKU = cell.String()
		}

		cell, ok = row[parser.EAN]
		if ok {
			sku.EAN = cell.String()
		}

		cell, ok = row[parser.Title]
		if ok {
			sku.Title = cell.String()
		}

		sku.Condition = dao.Unchecked

		sku.Quantity = 1

		switch intersect {
		case parser.EAN:
			sku.UnitCostInCent, ok = unitPriceByIntersect[sku.EAN]
		case parser.SKU:
			sku.UnitCostInCent, ok = unitPriceByIntersect[sku.AvidesSKU]
		default:
			err = fmt.Errorf("get the case without intersected header")

			return
		}

		if !ok {
			log.WarnFmt("Can't match Overview and Packinglist sheets by `%s` for `%s`", intersect, e.filePath)

			return
		}

		list = append(list, sku)
	}

	return
}

func getIntersectedHeader(headers parser.Headers) (intersect parser.Header, ok bool) {
	for _, intersect = range headers {
		if intersect == parser.SKU || intersect == parser.EAN {
			ok = true

			break
		}
	}

	return
}
