package extractor

import (
	"fmt"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/parser"
	"github.com/leprosus/golang-log"
	"io"
)

type Extractor struct {
	parser *parser.Parser
}

func NewExtractor(reader io.Reader) (e *Extractor, err error) {
	e = &Extractor{}

	e.parser, err = parser.NewParser(reader)
	if err != nil {
		return
	}

	return
}

func (e *Extractor) Extract() (list dao.UnitList, err error) {
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

	err = fmt.Errorf("unexpected type of document")

	return
}

func (e *Extractor) extractPriceList() (list dao.UnitList, err error) {
	var data parser.Data
	data, _, err = e.parser.Parse("Packinglist")
	if err != nil {
		return
	}

	var (
		unit dao.Unit
		cell parser.Cell
		ok   bool
	)
	for _, row := range data {
		unit = dao.Unit{
			Condition: dao.Returned,
		}

		cell, ok = row[parser.ASIN]
		if !ok || cell.String() == "" {
			continue
		}
		unit.ASIN = dao.ASIN(cell.String())

		cell, ok = row[parser.LagerId]
		if ok {
			unit.WarehouseId = cell.String()
		}

		cell, ok = row[parser.SKU]
		if ok {
			unit.SKU = cell.String()
		}

		cell, ok = row[parser.EAN]
		if ok {
			unit.EAN = cell.String()
		}

		cell, ok = row[parser.Condition]
		if ok {
			unit.Condition = dao.ConvCondition(cell.String())
		}

		unit.Quantity = 1
		cell, ok = row[parser.Quantity]
		if ok {
			unit.Quantity, err = cell.UInt32()
			if err != nil {
				return
			}
		}

		cell, ok = row[parser.UnitPrice]
		if !ok {
			continue
		}
		unit.UnitCostInCent, err = cell.PriceInCent()
		if err != nil {
			return
		}

		if unit.Quantity > 1 {
			unit.UnitCostInCent /= unit.Quantity
		}

		list = append(list, unit)
	}

	return
}

func (e *Extractor) extractCollection() (list dao.UnitList, err error) {
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
		price uint32

		unitPriceByIntersect = map[string]uint32{}
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

	var unit dao.Unit
	for _, row := range data {
		unit = dao.Unit{
			Condition: dao.Returned,
		}

		cell, ok = row[parser.ASIN]
		if !ok || cell.String() == "" {
			continue
		}
		unit.ASIN = dao.ASIN(cell.String())

		cell, ok = row[parser.LagerId]
		if ok {
			unit.WarehouseId = cell.String()
		}

		cell, ok = row[parser.SKU]
		if ok {
			unit.SKU = cell.String()
		}

		cell, ok = row[parser.EAN]
		if ok {
			unit.EAN = cell.String()
		}

		cell, ok = row[parser.Condition]
		if ok {
			unit.Condition = dao.ConvCondition(cell.String())
		}

		unit.Quantity = 1
		cell, ok = row[parser.Quantity]
		if ok {
			unit.Quantity, err = cell.UInt32()
			if err != nil {
				return
			}
		}

		cell, ok = row[parser.UnitPrice]
		if ok {
			unit.UnitCostInCent, err = cell.PriceInCent()
			if err != nil {
				return
			}
		} else {
			switch intersect {
			case parser.EAN:
				unit.UnitCostInCent, ok = unitPriceByIntersect[unit.EAN]
			case parser.SKU:
				unit.UnitCostInCent, ok = unitPriceByIntersect[unit.SKU]
			default:
				err = fmt.Errorf("get the case without intersected header")

				return
			}
		}

		if !ok {
			log.WarnFmt("Can't match Overview and Packinglist sheets by `%s`", intersect)

			return
		}

		list = append(list, unit)
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
