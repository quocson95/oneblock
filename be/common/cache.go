package common

import (
	"bytes"
	"compress/gzip"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

var CacheDataPool, _ = lru.New[string, CacheData](2048)

var CacheUser, _ = lru.New[string, User](1024)

type CacheData struct {
	Data            []byte
	Mime            string
	Header          map[string]string
	ContentLength   int64
	ContentEncoding string
	CreateAt        time.Time
	InvalidAt       time.Time
}

func (c *CacheData) Compress() {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte(c.Data))
	w.Close()
	c.Data = b.Bytes()
	c.ContentLength = int64(b.Len())
	c.ContentEncoding = "gzip"
}

func isMimeImage(mime string) bool {
	if mime == "image/webp" {
		return true
	}
	if mime == "image/png" {
		return true
	}
	if mime == "image/jpeg" {
		return true
	}
	return false
}
