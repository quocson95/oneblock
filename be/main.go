package main

import (
	"be/api"
	"be/config"
	"be/sp500"
	"fmt"
	"net/http"

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
	api.DefaultS3Hepler.Init(config.GetConfig().S3Endpoint, "hcm", config.GetConfig().S3AccessKey, config.GetConfig().S3SecretKey)
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},                                                      // Allow all origins
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
	e.GET("/", func(c echo.Context) error {
		c.JSON(http.StatusOK, "ok")
		return nil
	})
	e.GET("/be/data/btc_gold_raw", api.BtcGold)
	e.GET("/be/data/m1", api.MoneySupplyM1)
	e.GET("/be/data/m2", api.MoneySupplyM2)
	e.GET("/be/data/money_supply", api.MoneySupplyAgress)
	e.GET("/be/data/btc_gold", api.BtcGoldAgressApi)
	e.GET("/be/data/sp500", sp500.Sp500)

	e.GET("/be/data/funding_market_core", api.FundingMarketCore)
	e.GET("/be/data/btc_holder", api.BtcHolder)

	e.GET("/be/data/btc_eth_static", api.BtcEthStatic)
	e.GET("/be/data/storage", api.StorageFile)

	e.GET("/be/data/eth_gas_history", api.EthGasHistory)
	e.GET("/be/data/usd_vnd", api.USDVNDRate)
}

func static(e *echo.Echo) {
	e.Static("/be/chart", "chart")
}
