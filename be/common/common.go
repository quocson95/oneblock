package common

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

const DefaultBucketMdx = "mdx"

var GetDB func() *gorm.DB

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

func Sha265Random() string {
	hash := sha256.New()
	hash.Write([]byte(time.Now().String()))

	// Get the hash result
	hashBytes := hash.Sum(nil)

	// Convert the hash bytes to a hexadecimal string
	hashString := hex.EncodeToString(hashBytes)
	return hashString
}

func QuickMd5(data []byte) string {
	// Create a new MD5 hash
	hash := md5.New()
	// Write data to the hash
	hash.Write(data)
	// Get the resulting hash as a byte slice
	hashInBytes := hash.Sum(nil)
	// Convert the byte slice to a hexadecimal string
	hashString := hex.EncodeToString(hashInBytes)
	return hashString
}

func GetWeekRange(t time.Time) (time.Time, time.Time) {
	// Get the weekday (0=Sunday, 1=Monday, ..., 6=Saturday)
	weekday := int(t.Weekday())

	// Adjust for start of the week (Monday)
	// If you want Sunday as the first day of the week, adjust as needed
	if weekday == 0 {
		weekday = 7
	}

	// Start of the week (Monday)
	startOfWeek := t.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)

	// End of the week (Sunday)
	endOfWeek := startOfWeek.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	return startOfWeek, endOfWeek
}
