package apiadmin

import (
	"be/api/general"
	"be/common"
	"be/database"
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type MdxAdminController struct {
	api.MdxController
}

func (m *MdxAdminController) Handler(g *echo.Group) {
	m.MdxController.Handler(g)
	g.PUT("/", m.UploadMdx)
}

func (m *MdxAdminController) List(c echo.Context) {
	limit := 100
	offset := 0
	ml := make([]common.Mdx, 0)
	if v, _ := strconv.Atoi(c.QueryParam("limit")); v > 0 {
		limit = v
	}
	if v, _ := strconv.Atoi(c.QueryParam("offset")); v > 0 {
		offset = v
	}
	database.DB.Model(new(common.Mdx)).Limit(limit).Offset(offset).Order("id DESC").Find(&ml)
	c.JSON(http.StatusOK, ml)
}

func (m *MdxAdminController) UploadMdx(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusOK, common.Mdx{Err: errors.New("read body failed")})
	}
	// buf := &bytes.Buffer{}
	// zw := gzip.NewWriter(buf)
	name := c.QueryParam("name")
	preSign, err := common.DefaultS3Hepler.PreSign(http.MethodPut, common.DefaultBucketMdx.String(), name)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("presign failed")
		// c.AbortWithError(http.StatusBadRequest, errors.New("presign failed"))
		return c.JSON(http.StatusOK, common.Mdx{Err: errors.New("presign failed")})
	}
	if len(name) == 0 {
		name = uuid.New().String()
	}
	preSign.Url, _ = url.PathUnescape(preSign.Url)
	client, err := http.NewRequest(http.MethodPut, preSign.Url, bytes.NewReader(body))
	if err != nil {
		zap.L().With(zap.Error(err)).Error("init put failed")
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("init put failed"))
	}
	// go func() {
	// 	zw.Name = name
	// 	zw.ModTime = time.Now()
	// 	zw.Write(body)
	// 	zw.Close()
	// }()
	client.Header.Set("Content-Type", c.Request().Header.Get("Content-Type"))
	client.ContentLength = int64(len(body))
	resp, err := common.DefaultHttpClient.Do(client)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("put failed")
		// c.AbortWithError(http.StatusBadRequest, errors.New("put failed"))
		return c.JSON(http.StatusOK, common.Mdx{Err: errors.New("put failed")})
	}
	if resp.StatusCode != 200 {
		bodyErr, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		zap.L().With(zap.String("url", preSign.Url)).With(zap.ByteString("body", bodyErr)).With(zap.Int("status", resp.StatusCode)).Error("put failed")
		// c.AbortWithError(http.StatusBadRequest, errors.New("put failed"))
		return c.JSON(http.StatusOK, common.Mdx{Err: errors.New("put failed")})
	}
	mdx := &common.Mdx{}
	mdx.GetByName(database.DB, name)
	mdx.Name = name
	mdx.DisplayName = c.QueryParam("dispName")
	mdx.MD5 = common.QuickMd5(body)
	mdx.UpdatedAt = time.Now()
	if mdx.ID == 0 {
		mdx.Insert(database.DB)
		s3Sync := &common.S3ObjectSync{
			Name:     name,
			Bucket:   common.DefaultBucketMdx.String(),
			Source_1: common.SourceS3CloudFy,
		}
		if err := s3Sync.Insert(); err != nil {
			zap.L().With(zap.Error(err)).Error("save s3 sync object failed")
		}
	} else {
		mdx.Update(database.DB, (mdx.ID), map[string]interface{}{"md5": mdx.MD5, "updated_at": mdx.UpdatedAt})
	}

	mdx.GetByName(database.DB, name)
	getPresign, _ := common.DefaultS3Hepler.PreSign(http.MethodGet, common.DefaultBucketMdx.String(), name)
	mdx.Url = getPresign.Url
	return c.JSON(http.StatusOK, mdx)
}
