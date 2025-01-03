package main

import (
	"be/api"
	apiadmin "be/api_admin"
	"be/common"
	"be/config"
	"be/database"
	"be/job"
	"be/security"
	"be/sp500"
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"time"

	"github.com/andybalholm/brotli"
	"github.com/andybalholm/brotli/matchfinder"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

var (
	Version   string
	BuildTime string
	Commit    string
)

func init() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.DisableCaller = false
	config.DisableStacktrace = true
	logger, _ := config.Build(zap.AddCaller())
	zap.ReplaceGlobals(logger)
	zap.L().With(zap.String("version", Version)).With(zap.String("BuildTime", BuildTime)).With(zap.String("Commit", Commit)).Info("build info")
}

func main() {
	go startMetricsHandler(81)
	config.LoadConfig("config.json")
	common.DefaultS3Hepler.Init(config.GetConfig().S3Endpoint, "hcm", config.GetConfig().S3AccessKey, config.GetConfig().S3SecretKey)
	// e := echo.New()
	// e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
	// 	AllowOrigins: []string{"*"},                                                      // Allow all origins
	// 	AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS}, // Allow all methods
	// 	AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAccept},
	// }))
	port := 80
	if err := database.InitDB(config.GetConfig().PostgressDsn); err != nil {
		panic(fmt.Errorf("init db failed: %s", err.Error()))
	}
	database.DB.AutoMigrate(new(common.Mdx), new(common.User), new(common.ICS), new(common.CrawlLog), new(common.S3ObjectSync))
	common.GetDB = func() *gorm.DB {
		return database.DB
	}
	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx
	go job.StartJobCrawl(ctx)
	go job.StartJobCompressImage(ctx)

	startServeAPI(port, func(router *gin.Engine) {
		// defer pprof.Register(router)
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
		new(api.AccountApi).Handler(router.Group("/be/account"))
		api.NewOath2Api(config.GetConfig().GoogleConsole).Handler(router.Group("/be/auth"))
		new(api.ICSAPi).Handler(router.Group("/be/ics"))

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

func startServeAPI(port int, handler func(router *gin.Engine), onErr func(err error), onDone func()) {
	trustOrigin := make(map[string]struct{})
	for _, s := range config.GetConfig().TrustOrigin {
		trustOrigin[s] = struct{}{}
	}
	zap.L().With(zap.Strings("origins", config.GetConfig().TrustOrigin)).Info("trust origin")
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"https://*on"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAccept, "user-agent", "referer", "Cookie", "Authorize"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			if _, exist := trustOrigin[origin]; exist {
				return true
			}
			// return exist
			zap.L().With(zap.String("origin", origin)).Error("reject origin")
			return false
		},
		MaxAge: 24 * time.Hour,
	}))
	router.Use(func(c *gin.Context) {
		if !shouldCompress(c.Request) {
			return
		}
		c.Header("Content-Encoding", "br")
		c.Header("Vary", "Accept-Encoding")
		brWriter := brotli.NewWriterV2(c.Writer, brotli.DefaultCompression)
		x := &CompressMidle{c.Writer, brWriter, 0}
		c.Writer = x
		defer func() {
			brWriter.Close()
			c.Header("Content-Length", strconv.Itoa(x.Length))
		}()
		c.Next()
	})
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

func startMetricsHandler(port int) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

type CompressMidle struct {
	gin.ResponseWriter
	writer *matchfinder.Writer
	Length int
}

func (g *CompressMidle) WriteString(s string) (int, error) {
	g.Length += len(s)
	return g.writer.Write([]byte(s))
}

func (g *CompressMidle) Write(data []byte) (int, error) {
	g.Length += len(data)
	return g.writer.Write(data)
}

func shouldCompress(req *http.Request) bool {
	if !strings.Contains(req.Header.Get("Accept-Encoding"), "br") {
		return false
	}
	if strings.Contains(req.URL.Path, "be/s3") {
		return false
	}
	if strings.Contains(req.URL.Path, "be/data/storage") {
		return false
	}
	if strings.Contains(req.URL.Path, "api/storage") {
		return false
	}

	extension := filepath.Ext(req.URL.Path)
	if len(extension) < 4 { // fast path
		return true
	}

	switch extension {
	case ".png", ".gif", ".jpeg", ".jpg", ".webp":
		return false
	default:
		return true
	}
}
