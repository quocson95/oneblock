package api

import (
	"be/common"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func FundingMarketCore(c echo.Context) error {
	fundingMarketCore := &common.FundingMarketCore{}
	fromTime := time.Now().Add(-4 * 365 * 24 * time.Hour)
	err := fundingMarketCore.Load(fromTime)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	correlationData := common.CorrelationData{}
	err = common.Load(common.BtcGold, func(data []byte) error {
		return json.Unmarshal(data, &correlationData)
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	btcGoldAgress := &common.BtcGoldAgress{}
	btcGoldAgress.Agresss(correlationData, fromTime)
	fundingMarketCore.BtcPrice = btcGoldAgress.BtcPrice
	c.JSON(http.StatusOK, fundingMarketCore)
	return nil
}
