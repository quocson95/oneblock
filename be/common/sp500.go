package common

import (
	"fmt"
	"math"
	"time"
)

type Sp500 struct {
	Data   [][]float64 `json:"data"`
	Events interface{} `json:"events"`
}

type Sp500Agress struct {
	Labels         []string  `json:"labels,omitempty"`
	Sp500          []float64 `json:"sp_500,omitempty"`
	BtcPrice       []float64 `json:"btc_price,omitempty"`
	BtcCorrelation []float64 `json:"btc_correlation,omitempty"`
	M2             []float64 `json:"m2,omitempty"`
}

func (s *Sp500Agress) Agress(sp500 Sp500, btc BtcGoldAgress, fromTime time.Time) {
	s.BtcPrice = make([]float64, 0)
	s.Labels = make([]string, 0)
	s.Sp500 = make([]float64, 0)
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

	filter := filterOb(sp500.Data)
	t := time.Now()
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
		s.Sp500 = append(s.Sp500, filter[k])
		s.Labels = append(s.Labels, fmt.Sprintf("%02d-%02d", timeCheck.Year(), timeCheck.Month()))
		timeCheck = timeCheck.AddDate(0, 1, 0)
	}
	s.BtcPrice = btc.BtcPrice
	s.BtcCorrelation = btc.BtcCorrelation
}
