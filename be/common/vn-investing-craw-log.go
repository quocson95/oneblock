package common

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
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

type VnInvestingCrawlLog struct {
	gorm.Model
	Data     string          `json:"-"`
	DataObjs []EconomicEvent `gorm:"-" json:"Data"`
}

func (v *VnInvestingCrawlLog) Insert() error {
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

func GetVnInvestingCrawlLogs(offset, limit int) ([]VnInvestingCrawlLog, error) {
	ml := make([]VnInvestingCrawlLog, 0, limit)
	tx := GetDB().Model(new(VnInvestingCrawlLog)).Offset(offset).Limit(limit).Order("created_at DESC").Find(&ml)
	for idx, v := range ml {
		v.DataObjs = make([]EconomicEvent, 0)
		json.Unmarshal([]byte(v.Data), &v.DataObjs)
		ml[idx] = v
	}
	return ml, tx.Error
}
