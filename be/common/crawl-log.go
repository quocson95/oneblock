package common

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type CrawlLogEventId string

const (
	CrawlLogEventIdInvestingCalendar  CrawlLogEventId = "crawl_investing_calendar"
	CrawlLogEventIdDataTradingBtcGold CrawlLogEventId = "crawl_data_trading_btc_gold"
	CrawlLogEventIdDataTradingSP500   CrawlLogEventId = "crawl_data_trading_sp500"
)

type EconomicEvent struct {
	DateTime string `json:"datetime"`
	TimeUnix int64  `json:"timeUnix"`
	Currency string `json:"currency"`
	Event    string `json:"event"`
	Actual   string `json:"actual"`
	Forecast string `json:"forecast"`
	Previous string `json:"previous"`
}

type CrawlLog struct {
	gorm.Model
	EventId  CrawlLogEventId `json:"craw_name,omitempty"`
	Data     string          `json:"-,omitempty"`
	DataObjs []EconomicEvent `gorm:"-" json:"Data,omitempty"`
}

func (v *CrawlLog) Insert() error {
	v.CreatedAt = time.Now()
	v.UpdatedAt = time.Now()
	v.ID = 0
	if len(v.Data) == 0 {
		data, _ := json.Marshal(v.DataObjs)
		v.Data = string(data)
	}
	tx := GetDB().Model(v).Create(v)
	return tx.Error
}

func GetCrawlLogs(offset, limit int) ([]CrawlLog, error) {
	ml := make([]CrawlLog, 0, limit)
	tx := GetDB().Model(new(CrawlLog)).Offset(offset).Limit(limit).Order("created_at DESC").Find(&ml)
	for idx, v := range ml {
		v.DataObjs = make([]EconomicEvent, 0)
		json.Unmarshal([]byte(v.Data), &v.DataObjs)
		ml[idx] = v
	}
	return ml, tx.Error
}
