package extractor

import (
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/parser"
	"github.com/call2mall/catalog/utils"
	log "github.com/leprosus/golang-log"
	"io"
	"os"
)

func ExtractData(dirPath string) (list dao.UnitList, err error) {
	var filesPaths []string
	filesPaths, err = utils.Walk(dirPath, "zip")
	if err != nil {
		return
	}

	var (
		ex      *parser.Extractor
		skuPart dao.UnitList
	)
	for _, filePath := range filesPaths {
		err = utils.Unzip(filePath, func(reader io.ReadCloser, fileName string) (err error) {
			ex, err = parser.NewExtractor(reader)
			if err != nil {
				return
			}

			skuPart, err = ex.Extract()
			if err != nil {
				log.WarnFmt("Can't extract data from `%s` file `%s`: %v", filePath, fileName, err)

				err = nil

				return
			}

			if len(skuPart) == 0 {
				return
			}

			list = append(list, skuPart...)

			return
		})
		if err != nil {
			return
		}

		err = os.Remove(filePath)
		if err != nil {
			return
		}
	}

	return
}
