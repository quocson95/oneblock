package common

import (
	"database/sql"
	"time"
)

type Mdx struct {
	ID        uint         `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt time.Time    `json:"created_at,omitempty"`
	UpdatedAt time.Time    `json:"updated_at,omitempty"`
	DeletedAt sql.NullTime `gorm:"index" json:"deleted_at,omitempty"`
	Name      string       `json:"name,omitempty"`
	Location  string       `json:"location,omitempty"`
}
