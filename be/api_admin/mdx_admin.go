package apiadmin

import (
	"be/api"
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

type MdxAdminController struct {
	api.MdxController
}

func (m *MdxAdminController) Handler(g gin.IRoutes) {
	m.MdxController.Handler(g)
	g.PUT("/", m.UploadMdx)
}

func (m *MdxAdminController) List(c *gin.Context) {
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

func (m *MdxAdminController) UploadMdx(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, common.Mdx{Err: errors.New("read body failed")})
		return
	}
	// buf := &bytes.Buffer{}
	// zw := gzip.NewWriter(buf)
	name := c.Query("name")
	preSign, err := api.DefaultS3Hepler.PreSign(http.MethodPut, common.DefaultBucketMdx, name)
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
	resp, err := common.DefaultHttpClient.Do(client)
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
	mdx.DisplayName = c.Query("dispName")
	mdx.MD5 = common.QuickMd5(body)
	mdx.UpdatedAt = time.Now()
	if mdx.ID == 0 {
		mdx.Insert(database.DB)
		s3Sync := &common.S3ObjectSync{
			Name:     name,
			Bucket:   common.DefaultBucketMdx,
			Source_1: common.SourceS3CloudFy,
		}
		if err := s3Sync.Insert(); err != nil {
			zap.L().With(zap.Error(err)).Error("save s3 sync object failed")
		}
	} else {
		mdx.Update(database.DB, (mdx.ID), map[string]interface{}{"md5": mdx.MD5, "updated_at": mdx.UpdatedAt})
	}

	mdx.GetByName(database.DB, name)
	getPresign, _ := api.DefaultS3Hepler.PreSign(http.MethodGet, common.DefaultBucketMdx, name)
	mdx.Url = getPresign.Url
	c.JSON(http.StatusOK, mdx)
}
