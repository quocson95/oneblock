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

type ResizeRequest struct {
	Key string
	W   int
}

var queueResizeImage = make(chan *ResizeRequest, 200)

func StartJobResizeImage(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				zap.L().Info("stop job resize image")
				return
			case req := <-queueResizeImage:
				key := req.Key
				v, exist := common.CacheDataPool.Get(key)
				if !exist {
					break
				}
				if v.Resize {
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

				v.Resize = true
				common.CacheDataPool.Add(key, v)
				// zap.L().With(zap.String("key", key)).With(zap.Duration("cost", time.Since(t))).Info("compress done")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

}

func AddCompressImage(key string, w int) {
	select {
	case queueResizeImage <- &ResizeRequest{
		Key: key,
		W:   w,
	}:
		zap.L().With(zap.String("key", key)).Info("add resize compress")
	default:
		zap.L().With(zap.String("key", key)).Info("queue resize full")
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
		resized := transform.Resize(img, w, int(float64(w)/ratio), transform.Lanczos)
		enc := png.Encode
		enc(imgResize, resized)
	} else {
		// no need resize
		imgResize = bytes.NewBuffer(data)
	}
	imgWebp := &bytes.Buffer{}
	err = webpbin.NewCWebP().Quality(100).Input(imgResize).Output(imgWebp).Run()
	if err != nil {
		return nil, "", err
	}
	return imgWebp.Bytes(), "image/webp", err
}
