package cache

import "sync"

type CacheDataType int

const (
	CacheDataTypeImage CacheDataType = iota
	CacheDataTypeUser  CacheDataType = iota
)

var CacheData = sync.Map{}

type ImageData struct {
	Data []byte
	Mime string
}
