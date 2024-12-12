package common

import "gorm.io/gorm"

type Mdx struct {
	gorm.Model
	Name     string
	Location string
}
