package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
	EthGasId            DataRawID = "eth_gas"
	USDVNDId            DataRawID = "usd_vnd"
)

func Load(id DataRawID, fn func(data []byte) error) error {
	data, err := os.ReadFile(fmt.Sprintf("raw_data/%s.json", id))
	if err != nil {
		return err
	}
	return fn(data)
}
func ReadLine(id DataRawID, fn func(line []string) error) error {
	file, err := os.Open(fmt.Sprintf("raw_data/%s.json", id))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ",")
		if fn(line) != nil {
			return err
		}
	}
	return scanner.Err()
}
