package common

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type ICS struct {
	gorm.Model
	Uid       string `gorm:"index:uidx_name,unique" json:"uid"`
	Name      string `json:"name,omitempty"`
	Desp      string `json:"desp,omitempty"`
	StartUnix int64  `json:"start_unix,omitempty"`
	EndUnix   int64  `json:"end_unix,omitempty"`
}

func (i *ICS) GetById(id uint) error {
	if id == 0 {
		return errors.New("id not allow less than zero")
	}
	tx := GetDB().Model(i).Where("id=?", id).First(i)
	return tx.Error
}

func InsertMultiICS(ml []ICS) error {
	tx := GetDB().CreateInBatches(ml, len(ml))
	return tx.Error
}

func (i *ICS) Insert() error {
	i.CreatedAt = time.Now()
	i.UpdatedAt = i.CreatedAt
	tx := GetDB().Create(i)
	return tx.Error
}

func GetICS(start, end time.Time, offset int, limit int) ([]ICS, error) {
	ml := make([]ICS, 0)
	if limit < 0 {
		return ml, nil
	}
	tx := GetDB().Model(new(ICS)).Where("deleted_at is null AND start_unix >= ? AND end_unix <= ?", start.Unix(), end.Unix()).Offset(offset).Limit(limit).Order("start_unix DESC").Find(&ml)
	return ml, tx.Error
}

func (i *ICS) Updates(changes map[string]interface{}) error {
	if len(changes) == 0 {
		return nil
	}
	if i.ID == 0 {
		return errors.New("id not allow less than zero")
	}
	changes["updated_at"] = time.Now()
	return GetDB().Model(i).Where("id=? and deleted_at is null", i.ID).Updates(changes).Error
}
