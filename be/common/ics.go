package common

import (
	"time"

	"gorm.io/gorm"
)

type ICS struct {
	gorm.Model
	Name      string `json:"name,omitempty"`
	Desp      string `json:"desp,omitempty"`
	StartUnix int64  `json:"start_unix,omitempty"`
	EndUnix   int64  `json:"end_unix,omitempty"`
}

func (i *ICS) Insert() error {
	i.CreatedAt = time.Now()
	i.UpdatedAt = i.CreatedAt
	tx := GetDB().Create(i)
	return tx.Error
}

func GetICS(offset int, limit int) ([]ICS, error) {
	ml := make([]ICS, 0)
	if limit < 0 {
		return ml, nil
	}
	tx := GetDB().Model(new(ICS)).Where("deleted_at is null").Offset(offset).Limit(limit).Order("start_unix DESC").Find(&ml)
	return ml, tx.Error
}
