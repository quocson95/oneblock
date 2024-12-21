package common

import (
	"database/sql"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TypeDoc int

const (
	TypeDocDraf                  = 0
	TypeDocPublish       TypeDoc = 1
	TypeDocAboutUsMember TypeDoc = 2
)

type Mdx struct {
	ID        uint         `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt time.Time    `json:"created_at,omitempty"`
	UpdatedAt time.Time    `json:"updated_at,omitempty"`
	DeletedAt sql.NullTime `gorm:"index" json:"deleted_at,omitempty"`
	Name      string       `gorm:"unique" json:"name,omitempty"`
	Url       string       `gorm:"-" json:"url,omitempty"`
	MD5       string       `gorm:"column:md5" json:"md5,omitempty"`
	Err       error        `gorm:"-" json:"err,omitempty"`
	Content   string       `gorm:"-" json:"content,omitempty"`
	// Publish   bool         `json:"publish,omitempty"`
	TypeDoc TypeDoc `json:"type_doc,omitempty"`
}

func (m *Mdx) Insert(db *gorm.DB) error {
	m.ID = 0
	m.CreatedAt = time.Now()
	m.UpdatedAt = m.CreatedAt
	tx := db.Create(m)
	return tx.Error
}

func GetListMdx(typeDoc int, offset, limit int) ([]Mdx, error) {
	ml := make([]Mdx, 0)
	whereQuery := `deleted_at is null`
	if typeDoc >= 0 {
		whereQuery += fmt.Sprintf(" AND type_doc=%d", typeDoc)
	}
	tx := GetDB().Model(new(Mdx)).Where(whereQuery)
	tx = tx.Limit(limit).Offset(offset).Order("id DESC").Find(&ml)
	return ml, tx.Error
}

func (m *Mdx) Update(db *gorm.DB, id uint, changes map[string]interface{}) error {
	if len(changes) == 0 {
		return nil
	}
	tx := db.Model(m).Where("id=?", id).Updates(changes)
	return tx.Error
}

func (m *Mdx) GetById(db *gorm.DB, uint int) error {
	tx := db.Model(m).Where("id=?", uint).First(m)
	return tx.Error
}

func (m *Mdx) GetByName(db *gorm.DB, name string) error {
	tx := db.Model(m).Where("name=?", name).First(m)
	return tx.Error
}
