package grabber

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/call2mall/catalog/amazon"
	"github.com/call2mall/catalog/chrome"
	"github.com/call2mall/catalog/dao"
	"github.com/call2mall/catalog/dowloader"
	"github.com/call2mall/catalog/grabber/extractor"
	. "github.com/call2mall/catalog/grabber/imap"
	"github.com/call2mall/catalog/messaging/imap"
	"github.com/disintegration/imaging"
	"github.com/leprosus/golang-config"
	"github.com/leprosus/golang-log"
	"github.com/pkg/errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"time"
)

func ExportUnits() (err error) {
	dstPath := config.Path("grabber.path")
	err = os.MkdirAll(dstPath, 0755)
	if err != nil {
		return
	}

	var email imap.Email
	email, err = LoadLatestEmail()
	if err != nil {
		return
	}

	if email.UID == 0 {
		return
	}

	srcUrl := ExtractWetransferUrl(email.Text)
	if len(srcUrl) == 0 {
		return
	}

	var ok bool
	ok, err = dao.IsProcessedWetransferUrl(srcUrl)
	if err != nil || ok {
		return
	}

	defer func() {
		if err == nil {
			err = dao.InsertWetransferUrl(email.UID, srcUrl)
			if err != nil {
				return
			}
		}
	}()

	b := chrome.NewBrowser()
	b.Timeout(2 * time.Minute)

	err = dowloader.DownloadFromWetransfer(srcUrl, dstPath, b)
	if err != nil {
		return
	}

	var newUnitList dao.UnitList
	newUnitList, err = extractor.ExtractData(dstPath)
	fmt.Println(len(newUnitList))
	if err != nil || len(newUnitList) == 0 {
		return
	}

	newASINList := newUnitList.ExtractASINList()

	var publishedASINList dao.ASINList
	publishedASINList, err = dao.GetPublishedASIN()
	if err != nil {
		return
	}

	addingASINList := publishedASINList.Diff(newASINList)
	err = addingASINList.Store()
	if err != nil {
		return
	}

	removingASINList := newASINList.Diff(publishedASINList)
	err = dao.RemoveUnitListByASINList(removingASINList)
	if err != nil {
		return
	}

	err = newUnitList.Store()
	if err != nil {
		return
	}

	err = newASINList.PushToGrabberQueue()
	if err != nil {
		return
	}

	log.DebugFmt("New %d units were stored successful", len(newUnitList))

	return
}

func RunGrabber() (err error) {
	go defrostQueue()

	var (
		asinList dao.ASINList
		asin     dao.ASIN
	)
	for {
		asinList, err = dao.PopFromGrabberQueue(1)
		if err != nil {
			log.CriticalFmt("Can't pop new ASIN to grabber its pages: %v", err)

			continue
		}

		if len(asinList) == 0 {
			time.Sleep(time.Minute)

			continue
		}

		for _, asin = range asinList {
			err = grabASIN(asin)
			if err != nil {
				log.ErrorFmt("Can't grab data for ASIN %v: %v", asin, err)

				continue
			}
		}
	}
}

func defrostQueue() {
	const minute = 60

	var err error
	for range time.NewTicker(time.Minute).C {
		err = dao.DefrostGrabberQueue(minute)
		if err != nil {
			log.CriticalFmt("Can't defrost ASIN from grabber queue: %v", err)

			continue
		}
	}
}

func grabASIN(asin dao.ASIN) (err error) {
	var (
		a    *amazon.Amazon
		meta amazon.Meta
		ok   bool

		props dao.ASINProps

		size    = config.UInt32("grabber.image.size")
		dirPath = config.Path("grabber.image.path")

		img      []byte
		filePath string
	)

	func() {
		var success = false

		defer func() {
			if !success {
				e := asin.MarkGrabberAs(dao.Fail)
				if e != nil {
					err = errors.Wrap(err, e.Error())

					log.Warn(err.Error())
				}
			}
		}()

		log.DebugFmt("It is searching page for ASIN `%s`", asin)

		a = amazon.NewAmazon()
		if err != nil {
			log.ErrorFmt("Can't init amazon builder: %v", err)

			return
		}

		meta, ok, err = a.Extract(asin)
		if !ok && err != nil {
			log.Error(err.Error())

			return
		}

		if !ok {
			log.WarnFmt("Can't parse ASIN `%s`", asin)

			return
		}

		props = dao.ASINProps{
			ASIN: asin,
			Image: dao.Image{
				Bytes: meta.Bytes,
			},
			ASINMeta: dao.ASINMeta{
				Url: meta.Url,
				Category: dao.Category{
					Name: meta.Category,
				},
				Title: meta.Title,
			},
		}

		err = props.Store()
		if err != nil {
			log.ErrorFmt("Can't store ASIN `%s`: %v", asin, err)

			return
		}

		img, err = prepareImage(props.Image.Bytes, size)
		if err != nil {
			log.ErrorFmt("Can't prepare image for ASIN `%s`: %v", asin, err)

			return
		}

		filePath = asin.FilePath(dirPath)

		err = os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			log.ErrorFmt("Can't create directory `%s` for ASIN `%s`: %v", filepath.Dir(filePath), asin, err)

			return
		}

		err = os.WriteFile(filePath, img, 0666)
		if err != nil {
			log.ErrorFmt("Can't write image `%s` for ASIN `%s`: %v", filePath, asin, err)

			return
		}

		err = asin.Publish()
		if err != nil {
			log.ErrorFmt("Can't publish ASIN `%s`: %v", asin, err)

			return
		}

		success = true

		err = asin.MarkGrabberAs(dao.Done)
		if err != nil {
			log.ErrorFmt("Can't set status as `done` of grabber queue task for ASIN `%s`: %v", asin, err)

			return
		}

		log.DebugFmt("Page for ASIN `%s` is handled", asin)
	}()

	return
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
