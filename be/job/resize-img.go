package job

import (
	"be/common"
	"bytes"
	"image"
	_ "image/jpeg"
	"image/png"
	"time"

	"github.com/anthonynsimon/bild/transform"
	"github.com/nickalie/go-webpbin"
	"go.uber.org/zap"
	_ "golang.org/x/image/webp"
)

func AddCompressImage(key string, w int) {
	v, exist := common.CacheDataPool.Get(key)
	if !exist {
		return
	}
	if v.Resized {
		return
	}
	if w > 0 && v.IsImage() {
		t := time.Now()
		if newData, mime, err := resizeImage(w, v.Data); err == nil {
			v.Data = newData
			v.Mime = mime
			v.ContentLength = int64(len(newData))
			zap.L().With(zap.String("key", key)).With(zap.Duration("cost", time.Since(t))).Info("resize done")
		} else {
			zap.L().With(zap.String("key", key)).With(zap.Error(err)).Error("resize failed")
		}
	}
	v.Resized = true
	common.CacheDataPool.Add(key, v)
}

func resizeImage(w int, data []byte) ([]byte, string, error) {
	reader := bytes.NewBuffer(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, "", err
	}
	size := img.Bounds().Size()
	imgBuf := &bytes.Buffer{}
	if w < size.X {
		ratio := float64(size.X) / float64(size.Y)
		imgResize := transform.Resize(img, w, int(float64(w)/ratio), transform.Lanczos)
		enc := png.Encode
		enc(imgBuf, imgResize)
	} else {
		// no need resize
		imgBuf = bytes.NewBuffer(data)
	}
	imgWebp := &bytes.Buffer{}
	err = webpbin.NewCWebP().Quality(100).Input(imgBuf).Output(imgWebp).Run()
	if err != nil {
		return nil, "", err
	}
	return imgWebp.Bytes(), "image/webp", err
}
