package api

import (
	"be/common"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func EthGasHistory(c echo.Context) error {
	ethWei := &common.EthGasWei{
		TimeUnix: make([]int, 0),
		Wei:      make([]int, 0),
		Labels:   make([]string, 0),
	}
	fromTime := time.Now().AddDate(-4, 0, 0)
	fromTimeUnix := fromTime.Unix()
	err := common.ReadLine(common.EthGasId, func(data []string) error {
		if len(data) < 3 {
			return nil
		}
		unix, _ := strconv.Atoi(data[1])
		if int64(unix) < fromTimeUnix {
			return nil
		}
		wei, _ := strconv.Atoi(data[2])
		ethWei.TimeUnix = append(ethWei.TimeUnix, unix)
		ethWei.Wei = append(ethWei.Wei, wei/1000000000)
		timeCheck := time.Unix(int64(unix), 0)
		label := fmt.Sprintf("%02d-%02d", timeCheck.Year(), timeCheck.Month())
		ethWei.Labels = append(ethWei.Labels, label)
		return nil
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return &echo.HTTPError{
			Code:    http.StatusBadRequest,
			Message: "load data error"}
	}
	btc := common.CorrelationData{}
	err = common.Load(common.BtcGold, func(data []byte) error {
		return json.Unmarshal(data, &btc)
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	holderBtc := &common.HolderBtc{}
	holderBtc.Load(fromTime)
	for _, v := range holderBtc.Prices {
		ethWei.BtcPrice = append(ethWei.BtcPrice, math.Round(v.V*100)/100)
	}
	// btcAgress := &common.BtcGoldAgress{}
	// btcAgress.Agresss(btc, fromTime)
	// // asc
	// priceBtc := make(map[int]float64)
	// for idx, price := range btcAgress.BtcPrice {
	// 	priceBtc[btcAgress.Unix[idx]] = price
	// }
	// sort.Slice(btcAgress.Unix, func(i, j int) bool {
	// 	a := btcAgress.Unix[i]
	// 	b := btcAgress.Unix[j]
	// 	return a < b
	// })
	// for _, unix := range ethWei.TimeUnix {
	// 	for _, btcUnix := range btcAgress.Unix {
	// 		if btcUnix >= unix {
	// 			ethWei.BtcPrice = append(ethWei.BtcPrice, priceBtc[(btcUnix)])
	// 			break
	// 		}
	// 	}
	// }
	return c.JSON(http.StatusOK, ethWei)
}
