package api

import (
	"be/api/payos"
	"be/config"
	"net/http"

	"github.com/labstack/echo/v4"
)

type NuoiToiController struct{}

func (m *NuoiToiController) Handler(g *echo.Group) {
	cfg := config.GetConfig().PayOS
	payos.NewPayOS(cfg.ClientID, cfg.ApiKey, cfg.ChecksumKey).Handler(g.Group("/payment"))

	g.GET("/son", m.NuoiSon)
}

func (m *NuoiToiController) NuoiSon(c echo.Context) error {
	newUrl := c.Request().URL
	newUrl.Path = "/be/nuoitoi/payment?id=2"
	newUrl.RawQuery = ""
	return c.Redirect(http.StatusFound, newUrl.String())
}
