package api

import (
	"be/common"
	"be/database"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MdxController struct{}

func (m *MdxController) Handler(g gin.IRoutes) {
	g.GET("", m.List)
	g.GET("/", m.List)
	g.GET("/:id", m.Get)
}

func (m *MdxController) List(c *gin.Context) {
	limit := 100
	offset := 0
	if v, _ := strconv.Atoi(c.Query("limit")); v > 0 {
		limit = v
	}
	if v, _ := strconv.Atoi(c.Query("offset")); v > 0 {
		offset = v
	}
	typeDocStr := c.DefaultQuery("type_doc", "1")
	typeDoc, _ := strconv.Atoi(typeDocStr)
	ml, _ := common.GetListMdx(typeDoc, offset, limit)
	c.JSON(http.StatusOK, ml)
}

func (m *MdxController) Get(c *gin.Context) {
	id := c.Param("id")
	v := common.Mdx{}
	database.DB.Model(new(common.Mdx)).Where("id=?", id).First(&v)
	preSign, err := DefaultS3Hepler.PreSign(http.MethodGet, common.DefaultBucketMdx, v.Name)
	if err != nil {
		zap.L().With(zap.String("id", id)).With(zap.String("name", v.Name)).With(zap.Error(err)).Error("presign failed")
		c.Abort()
		return
	}
	v.Url = preSign.Url
	if len(c.Query("loadContent")) > 0 {
		resp, cleanup, err := common.QuickGetHttp(preSign.Method, preSign.Url, nil)
		defer cleanup()
		if err == nil {
			content, _ := io.ReadAll(resp.Body)
			v.Content = string(content)

		}
	}
	c.JSON(http.StatusOK, v)

}
