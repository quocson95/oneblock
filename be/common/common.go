package common

import (
	"fmt"
	"os"
)

type DataRawID string

const (
	MoneyM1Id           DataRawID = "money_supply_m1"
	MoneyM2Id           DataRawID = "money_supply_m2"
	BtcGold             DataRawID = "btc_gold"
	SP500Id             DataRawID = "sp500"
	FundingMarketCoreId DataRawID = "funding_market_corelation"
	HolderBtcCountId    DataRawID = "holder_btc_count"
	HolderBtcPriceId    DataRawID = "holder_btc_price"
)

func Load(id DataRawID, fn func(data []byte) error) error {
	data, err := os.ReadFile(fmt.Sprintf("raw_data/%s.json", id))
	if err != nil {
		return err
	}
	return fn(data)
}
