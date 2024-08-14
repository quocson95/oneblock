package main

import (
	btcgold "be/btc_gold"
	moneysupply "be/money_supply"
	"be/sp500"
	"fmt"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.DisableCaller = false
	config.DisableStacktrace = true
	logger, _ := config.Build(zap.AddCaller())
	zap.ReplaceGlobals(logger)
}

func main() {
	e := echo.New()
	port := 8080
	zap.L().With(zap.Int("port", port)).Info("start server")
	api(e)
	static(e)
	e.Logger.Fatal(e.Start(fmt.Sprintf("localhost:%d", port)))
}

func api(e *echo.Echo) {
	e.GET("/btc_gold_raw", btcgold.BtcGold)
	e.GET("/m1", moneysupply.MoneySupplyM1)
	e.GET("/m2", moneysupply.MoneySupplyM2)
	e.GET("/money_supply", moneysupply.MoneySupplyAgress)
	e.GET("/btc_gold", btcgold.BtcGoldAgressApi)
	e.GET("/sp500", sp500.Sp500)
}

func static(e *echo.Echo) {
	e.Static("chart", "chart")
}
