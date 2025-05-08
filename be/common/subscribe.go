package common

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Subscribe struct {
	Id         uint           `gorm:"primarykey" json:"id,omitempty"`
	UpdatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt  time.Time      `json:"created_at,omitempty"`
	UserId     int            `gorm:"unique" json:"userId,omitempty"`
	PlanId     int            `json:"planId,omitempty"`
	Plan       *Plan          `gorm:"foreignKey:PlanId" json:"plan,omitempty"`
	ExpireUnix int64          `json:"expireUnix,omitempty"`
	ExpireDate string         `json:"expireDate,omitempty"`
	Active     bool           `json:"active,omitempty"`
}

func (s *Subscribe) Create() error {
	t := time.Unix(s.ExpireUnix, 0)
	s.ExpireDate = fmt.Sprintf("%2d/%2d/%4d", t.Day(), t.Month(), t.Year())
	return GetDB().Model(s).Create(s).Error
}

func (s *Subscribe) GetByUser(userId int) error {
	return GetDB().Model(s).Where("user_id=?", userId).Find(s).Error
}

func (s *Subscribe) Updates(changes map[string]interface{}) error {
	if len(changes) == 0 {
		return nil
	}
	if s.Id <= 0 {
		return errors.New("invalid id, id must larger than 0")
	}
	changes["update_at"] = time.Now()
	return GetDB().Model(s).Updates(changes).Error
}
func (s *Subscribe) Delete() error {
	if s.Id <= 0 {
		return errors.New("invalid id, id must larger than 0")
	}
	return GetDB().Delete(s, s.Id).Error
}

func GetAllSubscribeByUser(userId int, deleted bool, offset int, limit int) ([]Subscribe, error) {
	ml := make([]Subscribe, 0)
	tx := GetDB().Model(new(Subscribe)).Preload("Plan")
	args := make([]interface{}, 0)
	query := "1=1"
	if userId > 0 {
		query += "  AND user_id=?"
		args = append(args, userId)
	}
	if deleted {
		query += " AND deleted_at is not null"
	} else {
		query += " AND deleted_at is null"
	}
	tx.Where(query, args...).Offset(offset).Limit(limit)
	err := tx.Order("id DESC").Find(&ml).Error
	return ml, err
}
