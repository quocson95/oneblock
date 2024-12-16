package api

import (
	"be/cache"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TradingViewApi struct{}

func (m *TradingViewApi) Handler(g *gin.RouterGroup) {
	g.GET("/heatmap", m.SnapshotHeatmap)
}

func (m *TradingViewApi) SnapshotHeatmap(c *gin.Context) {
	data, ok := cache.CacheData.Load(cache.CacheDataTypeImage)
	if !ok {
		c.Abort()
		return
	}
	img, ok := data.(*cache.ImageData)
	if !ok {
		c.Abort()
		return
	}
	c.Data(http.StatusOK, img.Mime, img.Data)
}
