package job

import (
	"be/common"
	"bytes"
	"context"
	"image"
	_ "image/jpeg"
	"image/png"
	"time"

	"github.com/anthonynsimon/bild/transform"
	"github.com/nickalie/go-webpbin"
	"go.uber.org/zap"
	_ "golang.org/x/image/webp"
)

type CompressRequest struct {
	Key string
	W   int
}

var compressImageQueue = make(chan *CompressRequest, 200)

func StartJobCompressImage(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				zap.L().Info("stop job compress image")
				return
			case req := <-compressImageQueue:
				key := req.Key
				v, exist := common.CacheDataPool.Get(key)
				if !exist {
					break
				}
				if req.W > 0 && v.IsImage() {
					t := time.Now()
					if newData, mime, err := resizeImage(req.W, v.Data); err == nil {
						v.Data = newData
						v.Mime = mime
						v.ContentLength = int64(len(newData))
						zap.L().With(zap.String("key", key)).With(zap.Duration("cost", time.Since(t))).Info("resize done")
					} else {
						zap.L().With(zap.String("key", key)).With(zap.Error(err)).Error("resize failed")
					}
				}
				if v.IsCompress() {
					break
				}
				t := time.Now()
				v.Compress()
				common.CacheDataPool.Add(key, v)
				time.Sleep(500 * time.Millisecond)
				zap.L().With(zap.String("key", key)).With(zap.Duration("cost", time.Since(t))).Info("compress done")
			}
		}
	}()

}

func AddCompressImage(key string, w int) {
	select {
	case compressImageQueue <- &CompressRequest{
		Key: key,
		W:   w,
	}:
		zap.L().With(zap.String("key", key)).Info("add queue compress")
	default:
		zap.L().With(zap.String("key", key)).Info("queue compress full")
	}
}

func resizeImage(w int, data []byte) ([]byte, string, error) {
	reader := bytes.NewBuffer(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, "", err
	}
	size := img.Bounds().Size()
	imgResize := &bytes.Buffer{}
	if w < size.X {
		ratio := float64(size.X) / float64(size.Y)
		resized := transform.Resize(img, w, int(float64(w)/ratio), transform.Gaussian)
		enc := png.Encode
		enc(imgResize, resized)
	} else {
		// no need resize
		imgResize = bytes.NewBuffer(data)
	}
	imgWebp := &bytes.Buffer{}
	err = webpbin.NewCWebP().Quality(80).Input(imgResize).Output(imgWebp).Run()
	if err != nil {
		return nil, "", err
	}
	return imgWebp.Bytes(), "image/webp", err
}
