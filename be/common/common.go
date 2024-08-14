package common

import (
	"fmt"
	"os"
)

type MoneyId string

const (
	MoneyM1Id MoneyId = "money_supply_m1"
	MoneyM2Id MoneyId = "money_supply_m2"
	BtcGold   MoneyId = "btc_gold"
)

func Load(id MoneyId, fn func(data []byte) error) error {
	data, err := os.ReadFile(fmt.Sprintf("raw_data/%s.json", id))
	if err != nil {
		return err
	}
	return fn(data)
}
