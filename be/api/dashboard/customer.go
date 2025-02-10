package dashboard

import (
	"be/common"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Customer struct{}

func (m *Customer) Handler(g *echo.Group) {
	g.GET("", m.CustomerUsers)
	g.GET("/", m.CustomerUsers)
}

func (m *Customer) CustomerUsers(c echo.Context) error {
	ml, err := common.GetCustomers()
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get list customer failed")
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, ml)
}
