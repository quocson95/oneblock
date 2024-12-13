package api

import (
	"be/common"
	"be/database"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (m *MdxApi) Handler(g *gin.RouterGroup) {
	g.GET("", m.List)
	g.GET("/", m.List)
	g.PUT("/", m.UploadMdx)
	g.GET("/:id", m.Get)
	g.POST("/image", m.UploadImage)
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
	database.DB.Model(new(common.Mdx)).Limit(limit).Offset(offset).Find(&ml)
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

func (m *MdxApi) UploadMdx(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, common.Mdx{Err: errors.New("read body failed")})
		return
	}
	// buf := &bytes.Buffer{}
	// zw := gzip.NewWriter(buf)
	name := c.Query("name")
	preSign, err := DefaultS3Hepler.PreSign(http.MethodPut, defaultBucketMdx, name)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("presign failed")
		// c.AbortWithError(http.StatusBadRequest, errors.New("presign failed"))
		c.JSON(http.StatusOK, common.Mdx{Err: errors.New("presign failed")})
		return
	}
	if len(name) == 0 {
		name = uuid.New().String()
	}
	preSign.Url, _ = url.PathUnescape(preSign.Url)
	client, err := http.NewRequest(http.MethodPut, preSign.Url, bytes.NewReader(body))
	if err != nil {
		zap.L().With(zap.Error(err)).Error("init put failed")
		c.AbortWithError(http.StatusBadRequest, errors.New("init put failed"))
		c.JSON(http.StatusOK, common.Mdx{Err: errors.New("init put failed")})
		return
	}
	// go func() {
	// 	zw.Name = name
	// 	zw.ModTime = time.Now()
	// 	zw.Write(body)
	// 	zw.Close()
	// }()
	client.Header.Set("Content-Type", c.ContentType())
	client.ContentLength = int64(len(body))
	resp, err := defaultHttpClient.Do(client)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("put failed")
		// c.AbortWithError(http.StatusBadRequest, errors.New("put failed"))
		c.JSON(http.StatusOK, common.Mdx{Err: errors.New("put failed")})
		return
	}
	if resp.StatusCode != 200 {
		bodyErr, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		zap.L().With(zap.String("url", preSign.Url)).With(zap.ByteString("body", bodyErr)).With(zap.Int("status", resp.StatusCode)).Error("put failed")
		// c.AbortWithError(http.StatusBadRequest, errors.New("put failed"))
		c.JSON(http.StatusOK, common.Mdx{Err: errors.New("put failed")})
		return
	}
	mdx := &common.Mdx{}
	mdx.GetByName(database.DB, name)
	mdx.Name = name
	mdx.MD5 = quickMd5(body)
	mdx.UpdatedAt = time.Now()
	if mdx.ID == 0 {
		mdx.Insert(database.DB)
	} else {
		mdx.Update(database.DB, (mdx.ID), map[string]interface{}{"md5": mdx.MD5, "updated_at": mdx.UpdatedAt})
	}
	mdx.GetByName(database.DB, name)
	getPresign, _ := DefaultS3Hepler.PreSign(http.MethodGet, defaultBucketMdx, name)
	mdx.Url = getPresign.Url
	c.JSON(http.StatusOK, mdx)
}

func (m *MdxApi) UploadImage(c *gin.Context) {}

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
