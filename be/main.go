package main

import (
	"be/api"
	apiadmin "be/api_admin"
	"be/common"
	"be/config"
	"be/database"
	"be/security"
	"be/sp500"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
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
	// e := echo.New()
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: []string{"*"},                                                      // Allow all origins
	// 	AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS}, // Allow all methods
	// 	AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAccept},
	// }))
	port := 8080
	err := database.InitDB(config.GetConfig().PostgressDsn)
	database.DB.AutoMigrate(new(common.Mdx), new(common.User), new(common.ICS), new(common.VnInvestingCrawlLog))

	common.GetDB = func() *gorm.DB {
		return database.DB
	}
	if err != nil {
		panic(err)
	}
	// time.AfterFunc(2*time.Second, func() {
	// 	job.StartJobSnapshotTradingViewHeatmap()
	// })
	ctx, cancel := context.WithCancel(context.Background())
	go api.JobCrawAndImportEventInvestingCalendar(ctx)
	startServeAPI(port, func(router *gin.Engine) {
		router.Static("/be/static", "static")

		router.Static("/be/chart", "chart")
		router.GET("/be/data/btc_gold_raw", api.BtcGold)
		router.GET("/be/data/m1", api.MoneySupplyM1)
		router.GET("/be/data/m2", api.MoneySupplyM2)
		router.GET("/be/data/money_supply", api.MoneySupplyAgress)
		router.GET("/be/data/btc_gold", api.BtcGoldAgressApi)
		router.GET("/be/data/sp500", sp500.Sp500)

		router.GET("/be/data/funding_market_core", api.FundingMarketCore)
		router.GET("/be/data/btc_holder", api.BtcHolder)

		router.GET("/be/data/btc_eth_static", api.BtcEthStatic)
		// router.GET("/be/data/storage", api.StorageFile)
		// router.GET("/api/storage", api.StorageFile)
		// router.GET("/api/storage", )
		new(api.S3Storage).Handler(router.Group("/be/data/storage"))
		new(api.S3Storage).Handler(router.Group("/api/storage"))
		new(api.S3Storage).Handler(router.Group("/be/s3"))

		router.GET("/be/data/eth_gas_history", api.EthGasHistory)
		router.GET("/be/data/usd_vnd", api.USDVNDRate)
		new(api.MdxController).Handler(router.Group("/be/mdx"))
		new(api.TradingViewApi).Handler(router.Group("/be/tradingview"))
		new(api.AccountApi).Handler(router.Group("/be/account"))
		api.NewOath2Api(config.GetConfig().GoogleConsole).Handler(router.Group("/be/auth"))
		new(api.ICSAPi).Handler(router.Group("/be/ics"))
		new(api.VnInvestingCrawl).Handler(router.Group("/be/vn-investing"))

		//auth
		new(apiadmin.MdxAdminController).Handler(router.Group("/be/admin/mdx").Use(security.TokenAuthMiddleware(database.DB)))
		new(apiadmin.ICSAdminController).Handler(router.Group("/be/admin/ics").Use(security.TokenAuthMiddleware(database.DB)))

	}, func(err error) {
		// e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
		zap.L().With(zap.Int("port", port)).With(zap.Error(err)).Error("startServeAPI failed")
		cancel()
	},
		func() {
			cancel()
		})

}

var trustOrigin = map[string]struct{}{"https://oneblock.vn": {}, "https://dev.oneblock.vn": {}, "https://blog.oneblock.vn": {}}

func startServeAPI(port int, handler func(router *gin.Engine), onErr func(err error), onDone func()) {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"https://*on"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAccept, "user-agent", "referer", "Cookie", "Authorize"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			_, exist := trustOrigin[origin]
			return exist
		},
		MaxAge: 24 * time.Hour,
	}))
	if gin.Mode() == gin.ReleaseMode {
		router.SetTrustedProxies(nil)
	}
	handler(router)
	zap.L().With(zap.Int("port", port)).Info("start server")
	err := router.Run(fmt.Sprintf(":%d", port))
	if err != nil && onErr != nil {
		onErr(err)
		return
	}
	onDone()
}

func createDefaultUser() {
	emails := []string{
		"dangquocson1995@gmail.com",
		"nghuuloc512@gmail.com",
		"nguyentrungbmt17@gmail.com",
		"haotran1689@gmail.com",
		"dtoan.bui@gmail.com",
	}
	for _, email := range emails {
		u := common.User{
			UserName:     email,
			Email:        email,
			Role:         common.RoleUserAdmin,
			UsdtInWallet: 0,
		}
		if err := u.Create(); err != nil {
			zap.L().With(zap.Error(err)).With(zap.String("email", email)).Error("add user failed")
		}
	}
}
