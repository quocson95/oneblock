package common

import (
	"bytes"
	"compress/gzip"
	"time"

	"github.com/andybalholm/brotli"
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
	Resize          bool
}

func (c *CacheData) Compress() {
	c.CompressBroti()
	// c.CompressGzip()
}
func (c *CacheData) CompressGzip() {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write([]byte(c.Data))
	c.Data = b.Bytes()
	c.ContentLength = int64(b.Len())
	c.ContentEncoding = "gzip"
}

func (c *CacheData) CompressBroti() {
	b := &bytes.Buffer{}
	brWriter := brotli.NewWriterV2(b, brotli.DefaultCompression)
	brWriter.Write(c.Data)
	brWriter.Close()
	c.Data = b.Bytes()
	c.ContentLength = int64(b.Len())
	c.ContentEncoding = "br"
}

func (c *CacheData) IsCompress() bool {
	return len(c.ContentEncoding) > 0 || c.IsImage()
}

func (c *CacheData) IsImage() bool {
	return isMimeImage(c.Mime)
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

func GetUserInCache(email string) (User, error) {
	user, ok := CacheUser.Get(email)
	if ok {
		return user, nil
	}
	u := &User{}
	err := u.FindByEmail(email, []RoleUser{RoleUserAdmin, RoleUserManager, RoluserCustomer}, "Subscribe", "Subscribe.Plan")
	CacheUser.Add(email, *u)
	return *u, err
}
