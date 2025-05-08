package dashboard

import (
	"be/common"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type PaymentController struct{}

func (p *PaymentController) Handler(g *echo.Group) {
	g.GET("/plans", p.Plans)
}

func (p *PaymentController) Plans(c echo.Context) error {
	ml, err := common.GetAllPlans()
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get all plans failed")
		c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, ml)
}
