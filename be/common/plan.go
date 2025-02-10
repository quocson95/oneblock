package common

import "time"

type PlanRenewType int

const (
	PlanRenewTypeMonth PlanRenewType = 1
	PlanRenewTypeYear  PlanRenewType = 2
)

type Plan struct {
	Id             uint          `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt      time.Time     `json:"created_at,omitempty"`
	Name           string        `json:"name,omitempty"`
	Price          int           `json:"price,omitempty"`
	PriceDisp      string        `json:"priceDisp,omitempty"`
	Currency       string        `json:"currency,omitempty"`
	Desp           string        `json:"desp,omitempty"`
	DurationExtend string        `json:"durationExtend,omitempty"`
	PlanRenewType  PlanRenewType `json:"planRenewType,omitempty"`
}

func (p *Plan) Create() error {
	return GetDB().Model(p).Create(p).Error
}

func (p *Plan) Get() error {
	return GetDB().Model(p).First(p).Error
}
