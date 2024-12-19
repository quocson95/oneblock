package common

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type RoleUser int

const (
	RoleUserAdmin   = 1
	RoleUserManager = 2
	RoluserMemner   = 3
	// RolerUserGuest = 2
)

type User struct {
	gorm.Model
	UserName     string    `json:"user_name,omitempty"`
	Email        string    `json:"email,omitempty"`
	Role         RoleUser  `json:"role,omitempty"`
	UsdtInWallet int64     `json:"usdt_in_wallet,omitempty"` // usdt
	LastLogin    time.Time `json:"last_login,omitempty"`
}

func (u *User) Create() error {
	u.ID = 0
	return GetDB().Create(u).Error
}

func (u *User) Read() error {
	if u.ID <= 0 {
		return errors.New("invalid id")
	}
	tx := GetDB().Model(u).First(u)
	return tx.Error
}

func (u *User) FindByEmail(email string) error {
	if len(email) <= 0 {
		return errors.New("invalid email")
	}
	tx := GetDB().Model(u).Where("email=?", email).First(u)
	return tx.Error
}

func (u *User) Update(changes map[string]interface{}) error {
	delete(changes, "id")
	delete(changes, "Id")
	delete(changes, "ID")
	if len(changes) == 0 {
		return nil
	}
	tx := GetDB().Model(u).Where("id=?", u.ID).Updates(changes)
	return tx.Error
}
