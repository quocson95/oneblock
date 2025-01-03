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
	Source_1   SourceS3 `gorm:"index" json:"source1,omitempty"`
	Source_2   SourceS3 `gorm:"default(0);index" json:"source2,omitempty"`
	PresignUrl string   `gorm:"-" json:"presignUrl,omitempty"`
}

func (s *S3ObjectSync) Insert() error {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	tx := GetDB().Create(s)
	return tx.Error
}

func (s *S3ObjectSync) InsertOrUpdate() error {
	v := &S3ObjectSync{
		Name:   s.Name,
		Bucket: s.Bucket,
	}
	_ = GetDB().First(v).Error
	if v.ID > 0 {
		// exist, do update
		changes := make(map[string]interface{})
		if s.Source_1 != v.Source_1 {
			changes["source_1"] = s.Source_1
		}
		if s.Source_1 != v.Source_1 {
			changes["source_2"] = s.Source_2
		}
		if len(changes) == 0 {
			return nil
		}
		GetDB().Model(v).Updates(changes)
	}
	return s.Insert()
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
