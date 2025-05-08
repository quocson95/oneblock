package apiadmin

import (
	api "be/api/general"
	"be/common"
	"be/database"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	user, ok := c.Get("user").(*common.User)
	if !ok {
		return c.NoContent(http.StatusBadRequest)
	}
	var name string
	var data string
	var dispName string
	// formData.set("title",data.title);
	// formData.set("category",data.category);
	// formData.set("description",data.description);
	// formData.set("heroImage",data.heroImage);
	// formData.set("tags",data.tags.join(","));
	// formData.set("dispName", dispName);
	// formData.set("content", mdxEditorRef.current.getMarkdown())
	// formData.set("pubDate",new Date().toISOString());
	zap.L().With(zap.String("heroImage", c.FormValue("heroImage"))).Info("debug")
	if len(c.FormValue("heroImage")) > 0 {
		name = c.FormValue("name")
		mdx := &common.Mdx{
			Content: c.FormValue("content"),
			FrontMatter: common.FrontMatter{
				HeroImage:   c.FormValue("heroImage"),
				Category:    c.FormValue("category"),
				Description: c.FormValue("description"),
				PubDate:     c.FormValue("pubDate"),
				Tags:        strings.Split(c.FormValue("tags"), ","),
				Title:       c.FormValue("title"),
			},
		}
		mdx.MergeFrontMatter()
		data = mdx.Content
		zap.L().With(zap.String("data", data)).Info("xxxx")
	} else {
		body, _ := io.ReadAll(c.Request().Body)
		data = string(body)
		name = c.QueryParam("name")
	}

	// if err != nil {
	// 	return c.JSON(http.StatusOK, common.Mdx{Err: errors.New("read body failed")})
	// }
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
	client, err := http.NewRequest(http.MethodPut, preSign.Url, strings.NewReader(data))
	if err != nil {
		zap.L().With(zap.Error(err)).Error("init put failed")
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("init put failed"))
	}
	client.Header.Set("Content-Type", c.Request().Header.Get("Content-Type"))
	client.ContentLength = int64(len(data))
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
	mdx.DisplayName = dispName
	mdx.MD5 = common.QuickMd5([]byte(data))
	mdx.UpdatedAt = time.Now()
	mdx.TypeDoc = common.TypeDocPublish
	if mdx.ID == 0 {
		mdx.CreatedBy = user.Email
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
		changes := map[string]interface{}{"md5": mdx.MD5, "updated_at": mdx.UpdatedAt}
		changes["updated_by"] = user.Email
		changes["type_doc"] = mdx.TypeDoc
		if len(mdx.CreatedBy) == 0 {
			changes["created_by"] = user.Email
		}
		mdx.Update(database.DB, (mdx.ID), changes)
	}

	mdx.GetByName(database.DB, name)
	getPresign, _ := common.DefaultS3Hepler.PreSign(http.MethodGet, common.DefaultBucketMdx.String(), name)
	mdx.Url = getPresign.Url
	return c.JSON(http.StatusOK, mdx)
}
