package parser

import (
	"fmt"
	"github.com/call2mall/catalog/dao"
	"github.com/leprosus/golang-log"
	"io"
)

type Extractor struct {
	parser *Parser
}

func NewExtractor(reader io.Reader) (e *Extractor, err error) {
	e = &Extractor{}

	e.parser, err = NewParser(reader)
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

	var headers Headers
	_, headers, err = e.parser.Parse("Packinglist")
	if err != nil {
		return
	}

	if headers.Contain(LagerId, EAN, UnitPrice) {
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
	var data Data
	data, _, err = e.parser.Parse("Packinglist")
	if err != nil {
		return
	}

	var (
		unit dao.Unit
		cell Cell
		ok   bool
	)
	for _, row := range data {
		unit = dao.Unit{
			Condition: dao.Returned,
		}

		cell, ok = row[ASIN]
		if !ok || cell.String() == "" {
			continue
		}
		unit.ASIN = dao.ASIN(cell.String())

		cell, ok = row[LagerId]
		if ok {
			unit.WarehouseId = cell.String()
		}

		cell, ok = row[SKU]
		if ok {
			unit.SKU = cell.String()
		}

		cell, ok = row[EAN]
		if ok {
			unit.EAN = cell.String()
		}

		cell, ok = row[Condition]
		if ok {
			unit.Condition = dao.ConvCondition(cell.String())
		}

		unit.Quantity = 1
		cell, ok = row[Quantity]
		if ok {
			unit.Quantity, err = cell.UInt32()
			if err != nil {
				return
			}
		}

		cell, ok = row[UnitPrice]
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
		data    Data
		headers Headers

		intersect Header
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
		cell  Cell
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

		cell, ok = row[UnitPrice]
		if !ok {
			err = fmt.Errorf("there aren't `%s` on Overview sheet", UnitPrice)

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

		cell, ok = row[ASIN]
		if !ok || cell.String() == "" {
			continue
		}
		unit.ASIN = dao.ASIN(cell.String())

		cell, ok = row[LagerId]
		if ok {
			unit.WarehouseId = cell.String()
		}

		cell, ok = row[SKU]
		if ok {
			unit.SKU = cell.String()
		}

		cell, ok = row[EAN]
		if ok {
			unit.EAN = cell.String()
		}

		cell, ok = row[Condition]
		if ok {
			unit.Condition = dao.ConvCondition(cell.String())
		}

		unit.Quantity = 1
		cell, ok = row[Quantity]
		if ok {
			unit.Quantity, err = cell.UInt32()
			if err != nil {
				return
			}
		}

		cell, ok = row[UnitPrice]
		if ok {
			unit.UnitCostInCent, err = cell.PriceInCent()
			if err != nil {
				return
			}
		} else {
			switch intersect {
			case EAN:
				unit.UnitCostInCent, ok = unitPriceByIntersect[unit.EAN]
			case SKU:
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

func getIntersectedHeader(headers Headers) (intersect Header, ok bool) {
	for _, intersect = range headers {
		if intersect == SKU || intersect == EAN {
			ok = true

			break
		}
	}

	return
}
