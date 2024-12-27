package common

import (
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

var CacheImageData, _ = lru.New[string, ImageData](2048)

var CacheUser, _ = lru.New[string, User](1024)

type ImageData struct {
	Data          []byte
	Mime          string
	Header        map[string]string
	ContentLength int64
	CreateAt      time.Time
	InvalidAt     time.Time
}
