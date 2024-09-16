package api

import (
	"be/common"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func USDVNDRate(c echo.Context) error {
	usdVndRaw := common.UsdVnd{}
	common.Load(common.USDVNDId, func(data []byte) error {
		return json.Unmarshal(data, &usdVndRaw)
	})
	fromTime := time.Now().AddDate(-4, 0, 0)
	fromTimeUnix := fromTime.Unix()
	result := common.UsdVndBtc{}
	if len(usdVndRaw.Chart.Result) == 0 || len(usdVndRaw.Chart.Result[0].Indicators.Quote) == 0 {
		return c.JSON(http.StatusOK, result)
	}
	// maxTs := int64(0)
	holderBtc := &common.HolderBtc{}
	holderBtc.Load(fromTime)
	mapBtcPrice := make(map[int64]float64)
	for _, v := range holderBtc.Prices {
		if v.T < fromTimeUnix {
			break
		}
		// result.BtcPrice = append(result.BtcPrice, math.Round(v.V*100)/100)
		mapBtcPrice[(v.T)] = math.Round(v.V*100) / 100
	}
	for idx, ts := range usdVndRaw.Chart.Result[0].Timestamp {
		if ts < fromTimeUnix {
			continue
		}

		usdVnd := usdVndRaw.Chart.Result[0].Indicators.Adjclose[0].Adjclose[idx]
		if *usdVnd <= 1 {
			continue
		}
		// maxTs = ts
		result.Unix = append(result.Unix, int(ts))
		result.VndUsd = append(result.VndUsd, math.Round(*usdVnd*100)/100)
		timeCheck := time.Unix(int64(ts), 0)
		label := fmt.Sprintf("%02d-%02d", timeCheck.Year(), timeCheck.Month())
		result.Labels = append(result.Labels, label)
		btcPrice := mapBtcPrice[ts]
		if btcPrice < 10 {
			btcPrice = mapBtcPrice[ts-3600]
		}
		if btcPrice < 10 {
			btcPrice = mapBtcPrice[ts+3600]
		}

		result.BtcPrice = append(result.BtcPrice, btcPrice)
	}

	return c.JSON(http.StatusOK, result)
}
