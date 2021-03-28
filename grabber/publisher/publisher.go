package publisher

import (
	"bufio"
	"bytes"
	"github.com/call2mall/catalog/dao"
	"github.com/disintegration/imaging"
	config "github.com/leprosus/golang-config"
	log "github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var header = http.Header{}

func init() {
	header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
}

func RunPublisher(threads uint) (err error) {
	path := config.Path("publisher.path")
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return
	}

	go defrostQueue()

	var ch chan dao.ASIN
	ch, err = runThreads(threads)
	if err != nil {
		return
	}

	var (
		asinList dao.ASINList
		asin     dao.ASIN
	)
	for {
		asinList, err = dao.PopFromPublisher(threads)
		if err != nil {
			log.CriticalFmt("Can't pop new ASIN to publish: %v", err)

			continue
		}

		if len(asinList) == 0 {
			time.Sleep(time.Minute)

			continue
		}

		for _, asin = range asinList {
			ch <- asin
		}
	}
}

func defrostQueue() {
	const minute = 60

	var err error
	for range time.NewTicker(time.Minute).C {
		err = dao.DefrostPublisher(minute)
		if err != nil {
			log.CriticalFmt("Can't defrost ASIN from publisher queue: %v", err)

			continue
		}
	}
}

func runThreads(threads uint) (ch chan dao.ASIN, err error) {
	ch = make(chan dao.ASIN, threads)

	var i uint
	for i = 0; i < threads; i++ {
		go publishProps(ch)
	}

	return
}

func publishProps(ch chan dao.ASIN) {
	var (
		err      error
		props    dao.ASINProps
		size     = config.UInt32("publisher.size")
		img      []byte
		dirPath  = config.Path("publisher.path")
		filePath string
	)

	for asin := range ch {
		func() {
			var success = false

			defer func() {
				if !success {
					e := asin.MarkPublisherAs(dao.Fail)
					if e != nil {
						err = errors.Wrap(err, e.Error())

						log.Warn(err.Error())
					}
				}
			}()

			log.DebugFmt("It is publishing data for ASIN `%s`", asin)

			props, err = dao.GetProps(asin)
			if err != nil {
				return
			}

			if props.Category.CatalogCategoryId == 0 {
				return
			}

			if len(props.Title) == 0 {
				return
			}

			if len(props.ImageName) == 0 {
				return
			}

			img, err = prepareImage(props.Image.Bytes, size)
			if err != nil {
				return
			}

			filePath = asin.FilePath(dirPath)

			err = os.MkdirAll(filepath.Dir(filePath), 0755)
			if err != nil {
				return
			}

			err = os.WriteFile(filePath, img, 0666)
			if err != nil {
				return
			}

			err = asin.Publish()
			if err != nil {
				return
			}

			success = true

			err = asin.MarkPublisherAs(dao.Done)
			if err != nil {
				log.ErrorFmt("Can't set status as `done` of publisher queue task for ASIN `%s`: %v", asin, err)

				return
			}

			log.DebugFmt("Unit ASIN `%s` is published", asin)
		}()
	}
}

func prepareImage(in []byte, size uint32) (out []byte, err error) {
	var src image.Image
	src, err = jpeg.Decode(bytes.NewReader(in))
	if err != nil {
		return
	}

	var rect image.Rectangle
	if src.Bounds().Max.X > src.Bounds().Max.Y {
		rect = image.Rect(0, 0, src.Bounds().Max.X, src.Bounds().Max.X)
	} else {
		rect = image.Rect(0, 0, src.Bounds().Max.Y, src.Bounds().Max.Y)
	}

	dst := image.NewNRGBA(rect)
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	draw.Draw(dst, rect, &image.Uniform{C: white}, image.Point{X: 0, Y: 0}, draw.Src)

	if src.Bounds().Max.X > src.Bounds().Max.Y {
		rect.Min.Y = (src.Bounds().Max.X - src.Bounds().Max.Y) / 2
	} else {
		rect.Min.X = (src.Bounds().Max.Y - src.Bounds().Max.X) / 2
	}

	draw.Draw(dst, rect, src, image.Point{X: 0, Y: 0}, draw.Src)

	dst = imaging.Resize(dst, int(size), int(size), imaging.Lanczos)

	var buf bytes.Buffer
	err = jpeg.Encode(bufio.NewWriter(&buf), dst, &jpeg.Options{Quality: 90})
	if err != nil {
		return
	}

	out = buf.Bytes()

	return
}
