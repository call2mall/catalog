package parser

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

type Cell struct {
	value     string
	imageName string
	bs        []byte
}

func (c Cell) String() string {
	return c.value
}

func (c Cell) Int64() (i64 int64, err error) {
	if c.value == "" {
		return
	}

	i64, err = strconv.ParseInt(c.value, 10, 64)

	return
}

func (c Cell) UInt64() (ui64 uint64, err error) {
	if c.value == "" {
		return
	}

	ui64, err = strconv.ParseUint(c.value, 10, 64)

	return
}

func (c Cell) Int32() (i32 int32, err error) {
	if c.value == "" {
		return
	}

	var i64 int64
	i64, err = strconv.ParseInt(c.value, 10, 32)
	if err != nil {
		return
	}

	i32 = int32(i64)

	return
}

func (c Cell) UInt32() (ui32 uint32, err error) {
	if c.value == "" {
		return
	}

	var ui64 uint64
	ui64, err = strconv.ParseUint(c.value, 10, 32)
	if err != nil {
		return
	}

	ui32 = uint32(ui64)

	return
}

func (c Cell) Uint() (ui64 uint64, err error) {
	if c.value == "" {
		return
	}

	ui64, err = strconv.ParseUint(c.value, 10, 64)

	return
}

func (c Cell) Float() (f64 float64, err error) {
	if c.value == "" {
		return
	}

	f64, err = strconv.ParseFloat(c.value, 64)

	return
}

func (c Cell) PriceInCent() (ui64 uint32, err error) {
	if c.value == "" {
		return
	}

	withCent := strings.Contains(c.value, ".")

	var f64 float64
	f64, err = c.Float()
	if err != nil {
		return
	}

	if withCent {
		f64 *= 100
	}

	ui64 = uint32(math.Ceil(f64))

	return
}

func (c Cell) HasImage() (ok bool) {
	return len(c.bs) > 0
}

func (c Cell) Image() (name string, bs []byte, err error) {
	if !c.HasImage() {
		err = errors.New("the cell doesn't contain image")

		return
	}

	name, bs = c.imageName, c.bs

	return
}

type Row map[Header]Cell

type Data []Row

type Header string

const (
	ASIN        Header = "asin"
	Brand       Header = "brand"
	Category    Header = "category"
	SubCategory Header = "subcategory"
	Color       Header = "color"
	Condition   Header = "condition"
	Quantity    Header = "quantity"
	Description Header = "description"
	Title       Header = "title"
	Image       Header = "image"
	UnitPrice   Header = "unit price"
	RetailPrice Header = "retail price"
	EAN         Header = "ean"
	TotalPrice  Header = "total price"
	Gender      Header = "gender"
	LagerId     Header = "lager id"
	Land        Header = "land"
	Size        Header = "size"
	SKU         Header = "sku"
	Units       Header = "units"
	Weight      Header = "weight"
	Type        Header = "type"
)

type Headers map[int]Header

func (h Headers) Contain(headers ...Header) (ok bool) {
	var (
		count, total int

		h1, h2 Header
	)

	total = len(headers)

	for _, h1 = range h {
		for _, h2 = range headers {
			if h1 == h2 {
				count++
			}
		}
	}

	ok = count == total

	return
}
