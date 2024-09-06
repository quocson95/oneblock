package common

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"
)

type FundingMarketCoreData struct {
	Data Data `json:"data"`
}

type Data struct {
	GetExecution GetExecution `json:"get_execution"`
}

type GetExecution struct {
	ExecutionQueued    interface{}        `json:"execution_queued"`
	ExecutionRunning   interface{}        `json:"execution_running"`
	ExecutionSucceeded ExecutionSucceeded `json:"execution_succeeded"`
	ExecutionFailed    interface{}        `json:"execution_failed"`
}

type ExecutionSucceeded struct {
	ExecutionID               string    `json:"execution_id"`
	RuntimeSeconds            int64     `json:"runtime_seconds"`
	GeneratedAt               time.Time `json:"generated_at"`
	Columns                   []string  `json:"columns"`
	Data                      []Datum   `json:"data"`
	RequestMaxResultSizeBytes int64     `json:"request_max_result_size_bytes"`
}

type Datum struct {
	InvestmentDate     string  `json:"investment_date"`
	InvestmentUnix     int64   `json:"investment_unix"`
	AmountRaised       float64 `json:"amount_raised"`
	CumSumAmountRaised float64 `json:"cum_sum_amount_raised"`
}

type FundingMarketCore struct {
	Dates              []string  `json:"dates,omitempty"`
	AmountRaises       []float64 `json:"amountRaises,omitempty"`
	CumSumAmountRaises []float64 `json:"cumSumAmountRaises,omitempty"`
	BtcPrice           []float64 `json:"btc_price,omitempty"`
}

func (f *FundingMarketCore) Load(fromDate time.Time) error {
	d := &FundingMarketCoreData{}
	err := Load(FundingMarketCoreId, func(data []byte) error {
		return json.Unmarshal(data, &d)
	})
	data := d.Data.GetExecution.ExecutionSucceeded.Data

	// Parse the date string to a time.Time object
	f.Dates = make([]string, 0)
	f.AmountRaises = make([]float64, 0)
	f.AmountRaises = make([]float64, 0)
	for _, d := range data {
		arr := strings.Split(d.InvestmentDate, "-")
		if len(arr) < 3 {
			continue
		}
		year, _ := strconv.Atoi(arr[0])
		month, _ := strconv.Atoi(arr[1])
		day, _ := strconv.Atoi(arr[2])
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if fromDate.After(date) {
			continue
		}
		// d.InvestmentUnix = date.Unix()
		f.Dates = append(f.Dates, d.InvestmentDate)
		f.AmountRaises = append(f.AmountRaises, math.Round(d.AmountRaised*100)/100)
		f.CumSumAmountRaises = append(f.CumSumAmountRaises, math.Round(d.CumSumAmountRaised*100)/100)

	}
	return err
}
