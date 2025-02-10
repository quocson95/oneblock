package common

import (
	"database/sql"
	"errors"
	"strings"
	"time"
)

type CopyTradeOrderStatus int

const (
	CopyTradeOrderStatusUnknow = 0
	CopyTradeOrderStatusOpen   = 1
	CopyTradeOrderStatusClosed = 2
)

type CopyTradeOrder struct {
	ID            uint                 `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt     time.Time            `json:"-"`
	UpdatedAt     time.Time            `json:"-"`
	DeletedAt     sql.NullTime         `gorm:"index" json:"-,omitempty"`
	StatusOrder   CopyTradeOrderStatus `json:"statusOrder,omitempty"`
	PNL           int                  `json:"pnl,omitempty"`
	Leverage      int                  `json:"leverage,omitempty"`
	Margin        int                  `json:"margin,omitempty"`
	Long          int                  `json:"long,omitempty"`
	DateCloseUnix int                  `json:"dateCloseUnix,omitempty"`
	Year          int                  `json:"year,omitempty"`
	Month         int                  `json:"month,omitempty"`
	Week          int                  `json:"week,omitempty"`
}

func (c *CopyTradeOrder) Insert() error {
	c.UpdatedAt = time.Now()
	c.CreatedAt = time.Now()
	c.Year = c.UpdatedAt.Year()
	c.Month = int(c.UpdatedAt.Month())
	_, c.Week = c.UpdatedAt.ISOWeek()
	return GetDB().Model(c).Create(c).Error
}

func (c *CopyTradeOrder) Get() error {
	if c.ID <= 0 {
		return errors.New("invalid id")
	}
	return GetDB().Model(c).Where("id=?", c.ID).First(c).Error
}

func (c *CopyTradeOrder) Updates(changes map[string]interface{}) error {
	if len(changes) == 0 {
		return nil
	}
	norChanes := make(map[string]interface{})
	for k, v := range changes {
		norChanes[strings.ToLower(k)] = v
	}
	norChanes["updated_at"] = time.Now()
	return GetDB().Model(c).Updates(changes).Error
}

func (c *CopyTradeOrder) Delete() error {
	if c.ID == 0 {
		return nil
	}
	return GetDB().Delete(c, c.ID).Error
}

func GetAllCopyTradeOrder(statusOrder int, from, to time.Time, offset, limit int) ([]CopyTradeOrder, error) {
	ml := make([]CopyTradeOrder, 0)
	tx := GetDB().Model(new(CopyTradeOrder))
	query := ""
	args := make([]interface{}, 0)
	if statusOrder != CopyTradeOrderStatusUnknow {
		query = "status_order=?"
		args = append(args, statusOrder)
	}
	if from.Unix() > 0 {
		query += " date_close_unix >=? "
		args = append(args, from.Unix())
	}
	if to.Unix() > 0 {
		query += " and date_close_unix <= ?"
		args = append(args, to.Unix())
	}
	if len(args) > 0 {
		tx = tx.Where(query, args...)
	}
	err := tx.Offset(offset).Limit(limit).Order("date_close_unix DESC").Find(&ml).Error
	return ml, err
}
