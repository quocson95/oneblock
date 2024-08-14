package btcgold

import (
	"be/common"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Define the struct for the JSON data

func BtcGold(c echo.Context) error {
	data, err := os.ReadFile("raw_data/btc_gold.json")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "read file error")
	}
	correlationData := common.CorrelationData{}
	json.Unmarshal(data, &correlationData)
	return c.JSON(http.StatusOK, correlationData)
}

func BtcGoldAgressApi(e echo.Context) error {
	correlationData := common.CorrelationData{}
	err := common.Load(common.BtcGold, func(data []byte) error {
		return json.Unmarshal(data, &correlationData)
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	m := &common.BtcGoldAgress{}
	t := time.Now()
	fromTime := time.Date(t.Year()-6, t.Month(), 0, 0, 0, 0, 0, t.Location())
	m.Agresss(correlationData, fromTime)
	return e.JSON(http.StatusOK, m)
}
