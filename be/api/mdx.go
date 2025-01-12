package api

import (
	"be/common"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type MdxController struct{}

func (m *MdxController) Handler(g *echo.Group) {
	g.GET("", m.List)
	g.GET("/", m.List)
	g.GET("/:id", m.Get)
}

func (m *MdxController) List(c echo.Context) error {
	limit := 100
	offset := 0
	if v, _ := strconv.Atoi(c.QueryParam("limit")); v > 0 {
		limit = v
	}
	if v, _ := strconv.Atoi(c.QueryParam("offset")); v > 0 {
		offset = v
	}
	typeDocStr := c.QueryParam("type_doc")
	if len(typeDocStr) == 0 {
		typeDocStr = "1"
	}
	typeDoc, _ := strconv.Atoi(typeDocStr)
	ml, _ := common.GetListMdx(typeDoc, offset, limit)
	return c.JSON(http.StatusOK, ml)
}

func (m *MdxController) Get(c echo.Context) error {
	id := c.Param("id")
	v := common.Mdx{}
	idInt, _ := strconv.Atoi(id)
	err := v.GetById(common.GetDB(), idInt)
	if err != nil {
		zap.L().With(zap.Int("id", idInt)).With(zap.Error(err)).Error("get mdx by id failed")
		return c.NoContent(http.StatusBadRequest)
	}

	preSign, err := common.DefaultS3Hepler.PreSign(http.MethodGet, common.DefaultBucketMdx.String(), v.Name)
	if err != nil {
		zap.L().With(zap.String("id", id)).With(zap.String("name", v.Name)).With(zap.Error(err)).Error("presign failed")
		return c.NoContent(http.StatusOK)
	}
	v.Url = preSign.Url
	if len(c.QueryParam("loadContent")) > 0 {
		resp, cleanup, err := common.QuickGetHttp(preSign.Method, preSign.Url, nil)
		defer cleanup()
		if err == nil {
			content, _ := io.ReadAll(resp.Body)
			v.Content = string(content)
		}
	}
	return c.JSON(http.StatusOK, v)

}
