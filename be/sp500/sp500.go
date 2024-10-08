package sp500

import (
	"be/common"
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func Sp500(c echo.Context) error {
	sp500 := common.Sp500{}
	err := common.Load(common.SP500Id, func(data []byte) error {
		return json.Unmarshal(data, &sp500)
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	agress := &common.Sp500Agress{}
	btc := common.CorrelationData{}
	err = common.Load(common.BtcGold, func(data []byte) error {
		return json.Unmarshal(data, &btc)
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	fromTime := time.Now().AddDate(-6, 0, 0)
	btcAgress := &common.BtcGoldAgress{}
	btcAgress.Agresss(btc, fromTime)
	agress.Agress(sp500, *btcAgress, fromTime)
	m2 := common.Money{}
	err = common.Load(common.MoneyM2Id, func(data []byte) error {
		return json.Unmarshal(data, &m2)
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	mAgress := common.Agresss(common.Money{}, m2, btc, fromTime)
	agress.M2 = mAgress.M2
	return c.JSON(http.StatusOK, agress)
}
