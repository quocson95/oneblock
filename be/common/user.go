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
	RoluserMember   = 3
	// RolerUserGuest = 2
)

type User struct {
	gorm.Model
	UserName     string    `json:"user_name,omitempty"`
	Email        string    `gorm:"unique" json:"email,omitempty"`
	Role         RoleUser  `json:"role,omitempty"`
	UsdtInWallet int64     `json:"usdt_in_wallet,omitempty"` // usdt
	LastLogin    time.Time `json:"last_login,omitempty"`
	Picture      string    `json:"picture,omitempty"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) Create() error {
	u.ID = 0
	return GetDB().Debug().Create(u).Error
}

func (u *User) Read() error {
	if u.ID <= 0 {
		return errors.New("invalid id")
	}
	tx := GetDB().Model(u).First(u)
	return tx.Error
}

func (u *User) FindByEmail(email string, roles ...RoleUser) error {
	if len(email) <= 0 {
		return errors.New("invalid email")
	}
	query := "email=?"
	args := make([]interface{}, 0)
	args = append(args, email)
	if len(roles) > 0 {
		query += " and role IN ?"
		args = append(args, roles)
	}
	tx := GetDB().Model(u).Where(query, args...)
	tx = tx.First(u)
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
