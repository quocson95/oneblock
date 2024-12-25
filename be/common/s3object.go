package common

import (
	"time"

	"gorm.io/gorm"
)

type SourceS3 int

const (
	SourceS3Unknow     SourceS3 = 0
	SourceS3CloudFy    SourceS3 = 1
	SourceS3CloudFlare SourceS3 = 2
)

type S3ObjectSync struct {
	gorm.Model
	Name       string   `gorm:"index:idx_name_age,unique" json:"name,omitempty"`
	Bucket     string   `gorm:"index:idx_name_age,unique" json:"bucket,omitempty"`
	Source_1   SourceS3 `gorm:"index" json:"source_1,omitempty"`
	Source_2   SourceS3 `gorm:"default(0);index" json:"source_2,omitempty"`
	PresignUrl string   `gorm:"-" json:"presign_url,omitempty"`
}

func (s *S3ObjectSync) Insert() error {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	tx := GetDB().Create(s)
	return tx.Error
}

func GetS3ObjectsSync(bucket string, offset, limit int) ([]S3ObjectSync, error) {
	ml := make([]S3ObjectSync, 0)
	tx := GetDB().Model(new(S3ObjectSync))
	if len(bucket) > 0 {
		tx = tx.Where("bucket=?", bucket)
	}
	tx = tx.Order("id desc").Offset(offset).Limit(limit)
	tx.Find(&ml)
	return ml, tx.Error
}
