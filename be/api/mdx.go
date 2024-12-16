package api

import (
	"be/common"
	"be/database"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const defaultBucketMdx = "mdx"

var defaultHttpClient = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
	},
}

type MdxApi struct{}

func (m *MdxApi) Handler(g gin.IRoutes) {
	g.GET("", m.List)
	g.GET("/", m.List)
	g.GET("/:id", m.Get)
}

func (m *MdxApi) List(c *gin.Context) {
	limit := 100
	offset := 0
	ml := make([]common.Mdx, 0)
	if v, _ := strconv.Atoi(c.Query("limit")); v > 0 {
		limit = v
	}
	if v, _ := strconv.Atoi(c.Query("offset")); v > 0 {
		offset = v
	}
	database.DB.Model(new(common.Mdx)).Limit(limit).Offset(offset).Order("id DESC").Find(&ml)
	c.JSON(http.StatusOK, ml)
}

func (m *MdxApi) Get(c *gin.Context) {
	id := c.Param("id")
	v := common.Mdx{}
	database.DB.Model(new(common.Mdx)).Where("id=?", id).First(&v)
	preSign, err := DefaultS3Hepler.PreSign(http.MethodGet, defaultBucketMdx, v.Name)
	if err != nil {
		zap.L().With(zap.String("id", id)).With(zap.String("name", v.Name)).With(zap.Error(err)).Error("presign failed")
		c.Abort()
		return
	}
	v.Url = preSign.Url
	if len(c.Query("loadContent")) > 0 {
		resp, cleanup, err := quickGetHttp(preSign.Method, preSign.Url, nil)
		defer cleanup()
		if err == nil {
			content, _ := io.ReadAll(resp.Body)
			v.Content = string(content)

		}
	}
	c.JSON(http.StatusOK, v)

}

func quickGetHttp(method, url string, body io.Reader) (*http.Response, func(), error) {
	client, err := http.NewRequest(method, url, body)
	cleanup := func() {}
	if err != nil {
		return nil, cleanup, err
	}
	resp, err := defaultHttpClient.Do(client)
	if err != nil {
		return nil, cleanup, err
	}
	cleanup = func() {
		if resp.Body == nil {
			return
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
	return resp, cleanup, err
}

func quickMd5(data []byte) string {
	// Create a new MD5 hash
	hash := md5.New()
	// Write data to the hash
	hash.Write(data)
	// Get the resulting hash as a byte slice
	hashInBytes := hash.Sum(nil)
	// Convert the byte slice to a hexadecimal string
	hashString := hex.EncodeToString(hashInBytes)
	return hashString
}
