package cache

import (
	"sync"
	"time"
)

type CacheDataType int

const (
	CacheDataTypeImage CacheDataType = iota
	CacheDataTypeUser  CacheDataType = iota
)

var CacheData = sync.Map{}

type ImageData struct {
	Data          []byte
	Mime          string
	Header        map[string]string
	ContentLength int64
	CreateAt      time.Time
	InvalidAt     time.Time
}
