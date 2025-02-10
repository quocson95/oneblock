package common

import (
	"database/sql"
	"errors"
	"time"
)

type RoleUser int

const (
	RoleUserAdmin   = 1
	RoleUserManager = 2
	RoluserCustomer = 3
	// RolerUserGuest = 2
)

type User struct {
	ID            uint         `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt     time.Time    `json:"created_at,omitempty"`
	UpdatedAt     time.Time    `json:"updated_at,omitempty"`
	DeletedAt     sql.NullTime `gorm:"index" json:"-"`
	UserName      string       `json:"userName,omitempty"`
	Email         string       `gorm:"unique" json:"email,omitempty"`
	Role          RoleUser     `json:"role,omitempty"`
	UsdtInWallet  int64        `json:"usdtInWallet,omitempty"` // usdt
	LastLoginUnix int64        `json:"lastLoginUnix,omitempty"`
	Picture       string       `json:"picture,omitempty"`
	// SubscribeId   uint         `json:"subscribeId,omitempty"`
	Subscribe Subscribe `json:"subscribe,omitempty"`
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

func (u *User) FindByEmail(email string, roles []RoleUser, joins ...string) error {
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
	tx := GetDB().Model(u)
	for _, join := range joins {
		tx.Preload(join)
	}
	tx = tx.Where(query, args...).First(u)
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

func GetCustomers() ([]User, error) {
	ml := make([]User, 0)
	err := GetDB().Debug().Model(new(User)).Preload("Subscribe").Preload("Subscribe.Plan").Where("role=?", RoluserCustomer).Find(&ml).Error
	return ml, err
}
