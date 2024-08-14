package common

import (
	"fmt"
	"time"
)

type MoneyAgrees struct {
	M1             []float64 `json:"m1,omitempty"`
	M2             []float64 `json:"m2,omitempty"`
	Labels         []string  `json:"labels,omitempty"`
	BtcCorrelation []float64 `json:"btc_correlation,omitempty"`
}

func Agresss(m1, m2 Money, btc CorrelationData) *MoneyAgrees {
	v := &MoneyAgrees{}
	v.Agresss(m1, m2, btc)
	return v
}

func (m *MoneyAgrees) Agresss(m1, m2 Money, btc CorrelationData) {
	t := time.Now()
	fromTime := time.Date(t.Year()-2, t.Month(), 0, 0, 0, 0, 0, t.Location())

	// filter data
	filterOb := func(m Money) map[string]float64 {
		fromMonth := fromTime.Month()
		fromYear := fromTime.Year()
		if len(m.Observations) == 0 {
			return map[string]float64{}
		}
		filterMap := make(map[string]float64)
		for _, ob := range m.Observations[0] {
			if len(ob) == 0 {
				continue
			}
			if ob[0] < 0 {
				continue
			}
			t1 := time.UnixMilli(int64(ob[0]))
			if t1.Year() < fromYear {
				continue
			}
			if t1.Year() >= fromYear || t1.Month() >= fromMonth {
				filterMap[fmt.Sprintf("%d_%d", t1.Year(), t1.Month())] = ob[1]
			}
		}
		return filterMap
	}
	filterM1 := filterOb(m1)
	filterM2 := filterOb(m2)
	filterBtc := BtcGoldAgress{}
	filterBtc.Agresss(btc, fromTime)
	// months := []int{0, 1, 2, 3, 4, 5, 6}
	// m.M1 = make([]float64, len(months))
	// m.M2 = make([]float64, len(months))
	// m.Labels = make([]string, len(months))
	// lM1 := len(m1.Observations[0])
	// lM2 := len(m2.Observations[0])
	// for idx, month := range months {
	// 	m.Labels[idx] = fmt.Sprintf("%d", month)
	// 	v := m1.Observations[0]
	// 	m.M1[idx] = v[lM1-1-idx][1]
	// 	m.M2[idx] = m1.Observations[0][lM2-1-idx][1]
	// }
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
		m.M1 = append(m.M1, filterM1[k])
		m.M2 = append(m.M2, filterM2[k])
		m.Labels = append(m.Labels, fmt.Sprintf("%02d-%02d", timeCheck.Year(), timeCheck.Month()))
		timeCheck = timeCheck.AddDate(0, 1, 0)
	}
	m.BtcCorrelation = filterBtc.BtcCorrelation
}
