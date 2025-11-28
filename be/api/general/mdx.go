package api

import (
	"be/common"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type MdxController struct {
	mtPrefetchMdx sync.Mutex
}

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
	go func() {
		m.mtPrefetchMdx.Lock()

		g := errgroup.Group{}
		g.SetLimit(10)
		for _, v := range ml {
			idStr := strconv.FormatInt(int64(v.ID), 10)
			if common.MdxCache.Contains(idStr) {
				continue
			}
			g.Go(func() error {
				_, err := m.fecthMdx(idStr, true)
				logger := zap.L().With(zap.Error(err)).With(zap.String("id", idStr))
				if err != nil {
					logger.Error("[failed] prefetch mdx ")
				} else {
					logger.Info("[success] prefetch mdx")
				}
				return nil
			})
		}
		g.Wait()
		m.mtPrefetchMdx.Unlock()
	}()
	return c.JSON(http.StatusOK, ml)
}

func (m *MdxController) Get(c echo.Context) error {
	id := c.Param("id")
	loadContent := false
	if len(c.QueryParam("loadContent")) > 0 {
		loadContent = true
	}
	v, err := m.fecthMdx(id, loadContent)
	if err != nil {
		zap.L().With(zap.String("id", id)).With(zap.Error(err)).Error("get mdx by id failed")
		return c.NoContent(http.StatusOK)
	}

	return c.JSON(http.StatusOK, v)

}

func (m *MdxController) fecthMdx(id string, loadContent bool) (common.Mdx, error) {
	v := common.Mdx{}
	idInt, _ := strconv.Atoi(id)

	err := v.GetById(common.GetDB(), idInt)
	if err != nil {
		return v, err
	}
	if common.MdxCache.Contains(id) {
		data, exist := common.MdxCache.Get(id)
		if exist {
			return data, nil
		}
	}

	preSign, err := common.DefaultS3Hepler.PreSign(http.MethodGet, common.DefaultBucketMdx.String(), v.Name)
	if err != nil {
		// zap.L().With(zap.String("id", id)).With(zap.String("name", v.Name)).With(zap.Error(err)).Error("presign failed")
		// return c.NoContent(http.StatusOK)
		return v, errors.Join(err, errors.New("name "+v.Name))
	}
	v.Url = preSign.Url
	if loadContent {
		resp, cleanup, err := common.QuickGetHttp(preSign.Method, preSign.Url, nil)
		defer cleanup()
		if err == nil {
			content, _ := io.ReadAll(resp.Body)
			v.Content = string(content)
		}
	}
	v.GenFrontMatter()
	v.MergeFrontMatter()
	common.MdxCache.Add(id, v)
	return v, nil
}
