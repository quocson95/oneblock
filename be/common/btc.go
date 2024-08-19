package common

import (
	"fmt"
	"math"
	"time"
)

type BtcCurrentPrice struct {
	Symbol         string `json:"symbol"`
	Price          string `json:"price"`
	PriceChange1H  string `json:"priceChange1H"`
	PriceChange24H string `json:"priceChange24H"`
}
type BtcPrice24H struct {
	Prices       [][]float64 `json:"prices"`
	MarketCaps   [][]float64 `json:"market_caps"`
	TotalVolumes [][]float64 `json:"total_volumes"`
}

type CorrelationData struct {
	BtcGoldCorrelation [][]float64 `json:"btc_gold_correlation"`
	GoldPriceUsd       [][]float64 `json:"gold_price_usd"`
	BtcDailyPrice      [][]float64 `json:"btc_daily_price"`
}
type BtcGoldAgress struct {
	Labels         []string  `json:"labels,omitempty"`
	BtcCorrelation []float64 `json:"btc_correlation,omitempty"`
	GoldPrice      []float64 `json:"gold_price,omitempty"`
	BtcPrice       []float64 `json:"btc_price,omitempty"`
}

func (m *BtcGoldAgress) Agresss(m1 CorrelationData, from time.Time) {
	t := time.Now()
	// fromTime := time.Date(t.Year()-2, t.Month(), 0, 0, 0, 0, 0, t.Location())
	fromTime := from

	// filter data
	filterOb := func(ml [][]float64) map[string]float64 {
		fromMonth := fromTime.Month()
		fromYear := fromTime.Year()
		if len(ml) == 0 {
			return map[string]float64{}
		}
		filterMap := make(map[string]float64)
		for _, ob := range ml {
			if len(ob) < 2 {
				continue
			}
			ts, val := ob[0], ob[1]
			t1 := time.UnixMilli(int64(ts))
			if t1.Year() < fromYear {
				continue
			}
			if t1.Year() >= fromYear || t1.Month() >= fromMonth {
				filterMap[fmt.Sprintf("%d_%d", t1.Year(), t1.Month())] = math.Round(val*100) / 100
			}
		}
		return filterMap
	}
	btcCorr := filterOb(m1.BtcGoldCorrelation)
	goldCorr := filterOb(m1.GoldPriceUsd)
	btcPrice := filterOb(m1.BtcDailyPrice)
	yearNow := t.Year()
	monthNow := t.Month()
	timeCheck := fromTime
	for {
		if timeCheck.Year() > yearNow {
			break
		}
		if timeCheck.Year() == yearNow && timeCheck.Month() > monthNow {
			break
		}
		k := fmt.Sprintf("%d_%d", timeCheck.Year(), timeCheck.Month())
		m.BtcCorrelation = append(m.BtcCorrelation, btcCorr[k])
		m.GoldPrice = append(m.GoldPrice, goldCorr[k])
		m.BtcPrice = append(m.BtcPrice, btcPrice[k])
		m.Labels = append(m.Labels, fmt.Sprintf("%02d-%02d", timeCheck.Year(), timeCheck.Month()))
		timeCheck = timeCheck.AddDate(0, 1, 0)
	}

}
