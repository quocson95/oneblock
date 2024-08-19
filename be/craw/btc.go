package craw

import (
	"be/common"
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type crawService struct {
	btcPrice *common.BtcCurrentPrice
}

var CrawService = NewCrawService()

func NewCrawService() *crawService {
	c := &crawService{}
	var err error
	c.btcPrice, err = BtcPriceSummary()
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get btc price summary error")
	}
	return c
}
func (c *crawService) GetBtcPrice() common.BtcCurrentPrice {
	btc := c.btcPrice
	return *btc
}

func init() {}

func BtcCurrentPrice() (*common.BtcCurrentPrice, error) {
	url := "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	btcPrice := &common.BtcCurrentPrice{}
	json.Unmarshal(body, &btcPrice)
	return btcPrice, nil
}

func BtcPrice24h() (*common.BtcPrice24H, error) {

	url := "https://api.coingecko.com/api/v3/coins/bitcoin/market_chart?vs_currency=usd&days=1"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	btcPrice24h := &common.BtcPrice24H{}
	json.Unmarshal(body, &btcPrice24h)
	return btcPrice24h, nil
}

func BtcPriceSummary() (*common.BtcCurrentPrice, error) {
	btcPrice, err := BtcCurrentPrice()
	if err != nil {
		return nil, err
	}
	btcChange, err := BtcPrice24h()
	if err != nil {
		return nil, err
	}
	// sort by timestamp desc
	sort.Slice(btcChange.Prices, func(i, j int) bool {
		a := btcChange.Prices[i][0]
		b := btcChange.Prices[j][0]
		return a > b
	})
	nowUnix := time.Now().Unix()
	btcCurPrice, _ := strconv.ParseFloat(btcPrice.Price, 64)
	for _, p := range btcChange.Prices {
		if nowUnix-int64(p[0]) >= 3600 && len(btcPrice.PriceChange1H) > 0 {
			priceChange := caclPercentChange(float64(p[1]), btcCurPrice)
			btcPrice.PriceChange1H = strconv.FormatFloat(priceChange, 'f', 2, 64)
		}
		if nowUnix-int64(p[0]) >= 86400 && len(btcPrice.PriceChange24H) > 0 {
			priceChange := caclPercentChange(float64(p[1]), btcCurPrice)
			btcPrice.PriceChange24H = strconv.FormatFloat(priceChange, 'f', 2, 64)
		}
	}
	return btcPrice, nil
}

func caclPercentChange(price1 float64, price2 float64) float64 {
	return (price2 - price1) / price1
}
