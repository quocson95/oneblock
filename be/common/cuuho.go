package common

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Cuuho struct {
	ID        int          `gorm:"primarykey" json:"id" `
	Content   string       `json:"content"`
	Phones    string       `json:"phones"`
	Area      string       `json:"area"`
	Status    string       `json:"status"`
	CreatedAt time.Time    `json:"created_at,omitempty"`
	UpdatedAt time.Time    `json:"updated_at,omitempty"`
	DeletedAt sql.NullTime `gorm:"index" json:"deleted_at,omitempty"`
}

func (m *Cuuho) Insert(db *gorm.DB) error {
	m.ID = 0
	m.CreatedAt = time.Now()
	m.UpdatedAt = m.CreatedAt
	tx := db.Create(m)
	return tx.Error
}

func (m *Cuuho) Update(db *gorm.DB, id uint, changes map[string]interface{}) error {
	if len(changes) == 0 {
		return nil
	}
	tx := db.Model(m).Where("id=?", id).Updates(changes)
	return tx.Error
}

func (m *Cuuho) GetById(db *gorm.DB, uint int) error {
	tx := db.Model(m).Where("id=?", uint).First(m)
	return tx.Error
}

func GetAllCuuho(offset, limit int) ([]Cuuho, error) {
	ml := make([]Cuuho, 0)
	whereQuery := `deleted_at is null`
	tx := GetDB().Model(new(Cuuho)).Where(whereQuery)
	tx = tx.Limit(limit).Offset(offset).Order("id DESC").Find(&ml)
	return ml, tx.Error
}
