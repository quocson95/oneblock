package api

import (
	"be/common"
	"be/database"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type MdxAdminApi struct {
	MdxApi
}

func (m *MdxAdminApi) Handler(g gin.IRoutes) {
	m.MdxApi.Handler(g)
	g.PUT("/", m.UploadMdx)
}

func (m *MdxAdminApi) List(c *gin.Context) {
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

func (m *MdxAdminApi) Get(c *gin.Context) {
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

func (m *MdxAdminApi) UploadMdx(c *gin.Context) {
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
