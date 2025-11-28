package apicuuho

import (
	"be/common"
	"be/database"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CuuhoController struct{}

func (m *CuuhoController) Handler(g *echo.Group) {
	g.GET("", m.List)
	g.GET("/:id", m.Get)
	g.POST("/", m.Add)
	g.PUT("/recuse/:id", m.recuse)
}

func (m *CuuhoController) List(c echo.Context) error {
	limit := 100
	offset := 0
	if v, _ := strconv.Atoi(c.QueryParam("limit")); v > 0 {
		limit = v
	}
	if v, _ := strconv.Atoi(c.QueryParam("offset")); v > 0 {
		offset = v
	}
	ml, _ := common.GetAllCuuho(offset, limit)
	return c.JSON(http.StatusOK, ml)
}

func (m *CuuhoController) Get(c echo.Context) error {
	id := c.Param("id")
	v := common.Cuuho{}
	idInt, _ := strconv.Atoi(id)
	err := v.GetById(common.GetDB(), idInt)
	if err != nil {
		zap.L().With(zap.Int("id", idInt)).With(zap.Error(err)).Error("get mdx by id failed")
		return c.NoContent(http.StatusOK)
	}
	return c.JSON(http.StatusOK, v)
}

func (m *CuuhoController) Add(c echo.Context) error {
	ml := make([]common.Cuuho, 0)
	err := c.Bind(&ml)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("parse body failed")
		return c.NoContent(http.StatusBadRequest)
	}
	for _, v := range ml {
		v.Insert(common.GetDB())
	}
	return c.NoContent(http.StatusOK)
}

func (m *CuuhoController) recuse(c echo.Context) error {
	id := c.Param("id")
	v := common.Cuuho{}
	idInt, _ := strconv.Atoi(id)
	err := v.GetById(common.GetDB(), idInt)
	if err != nil {
		zap.L().With(zap.Int("id", idInt)).With(zap.Error(err)).Error("get mdx by id failed")
		return c.NoContent(http.StatusOK)
	}
	changes := map[string]interface{}{"updated_at": time.Now()}
	v.Update(database.DB, uint(v.ID), changes)
	v.GetById(common.GetDB(), idInt)
	return c.JSON(http.StatusOK, v)
}
