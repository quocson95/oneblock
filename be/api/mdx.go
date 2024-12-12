package api

import (
	"be/common"
	"be/database"
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

func (m *MdxApi) Handler(g *gin.RouterGroup) {
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
	}
	resp, cleanup, err := quickGetHttp(preSign.Method, preSign.Url, nil)
	if err != nil {
		zap.L().With(zap.String("id", id)).With(zap.String("name", v.Name)).
			With(zap.String("url", preSign.Url)).With(zap.Error(err)).Error("get url failed")
		c.Abort()
	}
	defer cleanup()
	c.DataFromReader(http.StatusOK, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
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
