package dao

import (
	"strings"
)

type Condition string

const (
	Returned    Condition = "returned"
	Refurbished Condition = "refurbished"
	Unchecked   Condition = "unchecked"
	Defective   Condition = "defective"
	AWare       Condition = "a-ware"
	BWare       Condition = "b-ware"
	CWare       Condition = "c-ware"
)

func ConvCondition(origin string) (cond Condition) {
	switch strings.ToLower(origin) {
	case "retourware":
		cond = Returned
	case "refurbished":
		cond = Refurbished
	case "ungeprüfte retourware":
		cond = Unchecked
	case "defekt":
		cond = Defective
	case "a-ware":
		cond = AWare
	case "b-ware":
		cond = BWare
	case "c-ware":
		cond = CWare
	default:
		cond = Unchecked
	}

	return
}
