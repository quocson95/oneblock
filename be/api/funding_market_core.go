package api

import (
	"be/common"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func FundingMarketCore(c echo.Context) error {
	fundingMarketCore := &common.FundingMarketCore{}
	err := fundingMarketCore.Load(time.Now().Add(-4 * 365 * 24 * time.Hour))
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	c.JSON(http.StatusOK, fundingMarketCore)
	return nil
}
