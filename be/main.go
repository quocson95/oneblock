package main

import (
	"be/api"
	"be/config"
	"be/sp500"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	config.LoadConfig("config.json")
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins: []string{"*"}, // Allow all origins
        AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS}, // Allow all methods
        AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAccept},
    }))
	port := 8080
	zap.L().With(zap.Int("port", port)).Info("start server")
	apiHandler(e)
	static(e)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

func apiHandler(e *echo.Echo) {
	e.GET("/btc_gold_raw", api.BtcGold)
	e.GET("/m1", api.MoneySupplyM1)
	e.GET("/m2", api.MoneySupplyM2)
	e.GET("/money_supply", api.MoneySupplyAgress)
	e.GET("/btc_gold", api.BtcGoldAgressApi)
	e.GET("/sp500", sp500.Sp500)

	e.GET("/btc_eth_static", api.BtcEthStatic)
}

func static(e *echo.Echo) {
	e.Static("chart", "chart")
}
