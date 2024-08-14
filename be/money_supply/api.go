package moneysupply

import (
	"be/common"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

var moneySupply sync.Map
var mt = &sync.Mutex{}

type MoneyId string

const (
	MoneyM1Id MoneyId = "money_supply_m1"
	MoneyM2Id MoneyId = "money_supply_m2"
)

func load(id MoneyId) (common.Money, error) {
	mt.Lock()
	defer mt.Unlock()
	v, ok := moneySupply.Load(id)
	if ok {
		return v.(common.Money), nil
	}
	data, err := os.ReadFile(fmt.Sprintf("raw_data/%s.json", id))
	if err != nil {
		return common.Money{}, err
	}
	money := common.Money{}
	json.Unmarshal(data, &money)
	moneySupply.Store(id, money)
	return money, nil
}

func MoneySupplyM1(c echo.Context) error {
	money, err := load(MoneyM1Id)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	return c.JSON(http.StatusOK, money)
}

func MoneySupplyM2(c echo.Context) error {
	money, err := load(MoneyM2Id)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	return c.JSON(http.StatusOK, money)
}

func MoneySupplyAgress(c echo.Context) error {
	m1, err := load(MoneyM1Id)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	m2, err := load(MoneyM2Id)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	btc := common.CorrelationData{}
	err = common.Load(common.BtcGold, func(data []byte) error {
		return json.Unmarshal(data, &btc)
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, "load data error")
	}
	t := time.Now()
	fromTime := time.Date(t.Year()-6, t.Month(), 0, 0, 0, 0, 0, t.Location())
	mAgress := common.Agresss(m1, m2, btc, fromTime)
	return c.JSON(http.StatusOK, mAgress)
}
