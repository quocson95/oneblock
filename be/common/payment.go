package common

import (
	"database/sql"
	"time"

	"github.com/payOSHQ/payos-lib-golang"
)

type StatusPayment int

const (
	StatusPaymentUnknow      = 1
	StatusPaymentCreated     = 1
	StatusPaymentCancel      = 2
	StatusPaymentPaidSuccess = 3
	StatusPaymentPaidFailed  = 3
)

type Payment struct {
	ID              uint          `gorm:"primarykey" json:"id,omitempty"`
	Bin             string        `json:"bin,omitempty"`
	CreatedAt       time.Time     `json:"created_at,omitempty"`
	UpdatedAt       time.Time     `json:"updated_at,omitempty"`
	DeletedAt       sql.NullTime  `gorm:"index" json:"-,omitempty"`
	CustomerId      int           `json:"customerId,omitempty"`
	CustomerEmail   string        `json:"-"`
	OrderCode       int64         `json:"orderCode,omitempty"`
	Amount          int           `json:"amount,omitempty"`
	Description     string        `json:"description,omitempty"`
	AccountNumber   string        `json:"accountNumber,omitempty"`
	AccountName     string        `json:"accountName,omitempty"`
	Reference       string        `json:"reference,omitempty"`
	TransactionDate string        `json:"transactionDateTime,omitempty"`
	Currency        string        `json:"currency,omitempty"`
	PaymentLinkId   string        `json:"paymentLinkId,omitempty"`
	Code            string        `json:"code,omitempty"`
	Desc            string        `json:"desc,omitempty"`
	Status          StatusPayment `json:"status,omitempty"`
	QRCode          string        `json:"qrCode,omitempty"`
	QrLink          string        `gorm:"-" json:"qrLink,omitempty"`
	PlanId          int           `json:"planId,omitempty"`
	Plan            *Plan         `gorm:"foreignKey:PlanId" json:"plan,omitempty"`
	CheckoutUrl     string        `gorm:"-" json:"checkoutUrl,omitempty"`
	CdCheckoutBySec int64         `json:"cdCheckoutBySec,omitempty"`
}

func NewPayment(userID int, data *payos.CheckoutResponseDataType) *Payment {
	p := &Payment{
		CustomerId:    userID,
		OrderCode:     data.OrderCode,
		Amount:        data.Amount,
		Description:   data.Description,
		Status:        StatusPaymentCreated,
		AccountNumber: data.AccountNumber,
		AccountName:   data.AccountName,
		PaymentLinkId: data.PaymentLinkId,
		Currency:      data.Currency,
	}
	return p
}

func (p *Payment) Insert() error {
	return GetDB().Model(p).Create(p).Error
}

func (p *Payment) FindByPaymentLinkID(paymentLinkId string) error {
	return GetDB().Model(p).Where("payment_link_id=?", paymentLinkId).Preload("Plan").First(p).Error
}

func (p *Payment) FindByOrderCode(userId int, orderCode int64) error {
	if userId > 0 {
		return GetDB().Model(p).Where("customer_id=? and order_code=?", userId, orderCode).Preload("Plan").First(p).Error
	}
	return GetDB().Model(p).Where("order_code=?", orderCode).First(p).Error
}

func (p *Payment) Update(changes map[string]interface{}) error {
	if len(changes) == 0 {
		return nil
	}
	changes["updated_at"] = time.Now()
	return GetDB().Model(p).Updates(changes).Error
}
