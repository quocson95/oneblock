package common

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type TypeDoc int

var MdxCache, _ = lru.New[string, Mdx](1000)

const (
	TypeDocDraf                  = 0
	TypeDocPublish       TypeDoc = 1
	TypeDocAboutUsMember TypeDoc = 2
)

type FrontMatter struct {
	HeroImage   string   `yaml:"heroImage" json:"heroImage,omitempty"`
	Category    string   `yaml:"category" json:"category,omitempty"`
	Description string   `yaml:"description" json:"description,omitempty"`
	PubDate     string   `yaml:"pubDate" json:"pubate,omitempty"`
	Tags        []string `yaml:"tags" json:"tags,omitempty"`
	Title       string   `yaml:"title" json:"title,omitempty"`
}

type Mdx struct {
	ID          uint         `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt   time.Time    `json:"created_at,omitempty"`
	UpdatedAt   time.Time    `json:"updated_at,omitempty"`
	DeletedAt   sql.NullTime `gorm:"index" json:"deleted_at,omitempty"`
	Name        string       `gorm:"unique" json:"name,omitempty"`
	DisplayName string       `gorm:"default('')" json:"display_name,omitempty"`
	Url         string       `gorm:"-" json:"url,omitempty"`
	MD5         string       `gorm:"column:md5" json:"md5,omitempty"`
	Err         error        `gorm:"-" json:"err,omitempty"`
	FrontMatter FrontMatter  `gorm:"-" json:"frontMatter,omitempty"`
	Content     string       `gorm:"-" json:"content,omitempty"`
	// Publish   bool         `json:"publish,omitempty"`
	TypeDoc   TypeDoc `json:"type_doc,omitempty"`
	CreatedBy string  `json:"created_by,omitempty"`
	UpdatedBy string  `json:"updated_by,omitempty"`
}

func (m *Mdx) GenFrontMatter() error {
	if len(m.Content) == 0 {
		return nil
	}
	fm := strings.SplitN(string(m.Content), "---", 3)
	if len(fm) < 3 {
		return nil
	}
	var meta FrontMatter
	err := yaml.Unmarshal([]byte(fm[1]), &meta)
	if err != nil {
		return err
	}
	m.FrontMatter = meta
	return nil
}

func (m *Mdx) MergeFrontMatter() error {
	fm, _ := yaml.Marshal(m.FrontMatter)
	oriFm := strings.SplitN(string(m.Content), "---", 3)
	builder := strings.Builder{}

	if len(oriFm) < 3 {
		builder.WriteString("---\n")
		builder.WriteString(string(fm))
		builder.WriteString("---\n")
		builder.WriteString(m.Content)
	} else {
		builder.WriteString("---\n")
		builder.WriteString(string(fm))
		builder.WriteString("---\n")
		builder.WriteString(oriFm[2])
	}
	builder.WriteString("\n")
	m.Content = builder.String()
	return nil
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
	for idx, v := range ml {
		if len(v.DisplayName) == 0 {
			v.DisplayName = v.Name
		}
		v.GenFrontMatter()
		ml[idx] = v
	}
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
	if len(m.DisplayName) == 0 {
		m.DisplayName = m.Name
	}
	m.GenFrontMatter()
	return tx.Error
}

func (m *Mdx) GetByName(db *gorm.DB, name string) error {
	tx := db.Model(m).Where("name=?", name).First(m)
	if len(m.DisplayName) == 0 {
		m.DisplayName = m.Name
	}
	m.GenFrontMatter()
	return tx.Error
}
