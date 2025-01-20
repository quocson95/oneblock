package main

import (
	apiadmin "be/api/admin"
	"be/api/dashboard"
	api "be/api/general"
	"be/common"
	"be/config"
	"be/database"
	"be/job"
	"be/security"
	"be/sp500"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
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
	port := 8080
	if err := database.InitDB(config.GetConfig().PostgressDsn); err != nil {
		panic(fmt.Errorf("init db failed: %s", err.Error()))
	}
	database.DB.AutoMigrate(new(common.Mdx), new(common.User), new(common.ICS), new(common.CrawlLog), new(common.S3ObjectSync))
	common.GetDB = func() *gorm.DB {
		return database.DB
	}
	createDefaultUser()

	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx
	job.JobCrawlInvestingCalendar()
	go job.StartJobCrawl(ctx)
	go job.StartJobResizeImage(ctx)
	startEchoServeAPI(port, func(router *echo.Echo) {
		fnHello := func(c echo.Context) error {
			return c.HTML(http.StatusOK, "oneblock")
		}
		{
			router.GET("", fnHello)
			router.GET("/", fnHello)
			router.GET("/favicon.ico", func(c echo.Context) error {
				return c.File("favicon.svg")
			})
		}
		beRouter := router.Group("/be")
		{
			staticRouter := beRouter.Group("/static")
			staticRouter.Static("", "static")
			staticRouter.Static("/", "static")
		}
		beRouter.Static("/chart", "chart")
		{
			dataRouter := beRouter.Group("/data")
			dataRouter.GET("/btc_gold_raw", api.BtcGold)
			dataRouter.GET("/m1", api.MoneySupplyM1)
			dataRouter.GET("/m2", api.MoneySupplyM2)
			dataRouter.GET("/money_supply", api.MoneySupplyAgress)
			dataRouter.GET("/btc_gold", api.BtcGoldAgressApi)
			dataRouter.GET("/sp500", sp500.Sp500)
			dataRouter.GET("/funding_market_core", api.FundingMarketCore)
			dataRouter.GET("/btc_holder", api.BtcHolder)
			dataRouter.GET("/btc_eth_static", api.BtcEthStatic)

			new(api.S3Storage).Handler(router.Group("/storage"))
			new(api.S3Storage).Handler(router.Group("/api/storage"))
			new(api.S3Storage).Handler(router.Group("/be/s3"))

			dataRouter.GET("/eth_gas_history", api.EthGasHistory)
			router.GET("/usd_vnd", api.USDVNDRate)

		}
		new(api.MdxController).Handler(beRouter.Group("/mdx"))
		new(api.AccountApi).Handler(beRouter.Group("/account"))
		api.NewOath2Api(config.GetConfig().GoogleConsole, nil, nil).Handler(beRouter.Group("/auth"))
		new(api.ICSAPi).Handler(beRouter.Group("/ics"))

		//auth
		{
			adminRouter := beRouter.Group("/admin")
			adminRouter.Use(echojwt.WithConfig(echojwt.Config{
				// ...
				SigningKey:     []byte(security.SecretJwtAuth),
				SuccessHandler: security.SuccessHandlerUser(database.DB),
				// ContinueOnIgnoredError: false,
				ErrorHandler: func(c echo.Context, err error) error {
					zap.L().With(zap.Error(err)).Error("jwt handler error")
					return c.NoContent(http.StatusUnauthorized)
				},
				TokenLookup: "header:Authorization",
				// ...
			}))
			new(apiadmin.MdxAdminController).Handler(adminRouter.Group("/mdx"))
			new(apiadmin.ICSAdminController).Handler(adminRouter.Group("/ics"))
		}
		// dashboard
		{
			// be/dashboard/sso/google
			suffixSkips := []string{"/sso/google", "/sso/google/callback"}
			dashboardRouter := beRouter.Group("/dashboard")
			dashboardRouter.Use(echojwt.WithConfig(echojwt.Config{
				// ...
				SigningKey:     []byte(security.SecretJwtAuthDashboard),
				SuccessHandler: security.SuccessHandlerDashboardUser(database.DB),
				// ContinueOnIgnoredError: false,
				ErrorHandler: func(c echo.Context, err error) error {
					zap.L().With(zap.Error(err)).Error("jwt handler error")
					return c.NoContent(http.StatusUnauthorized)
				},
				TokenLookup: "header:Authorization",
				Skipper: func(c echo.Context) bool {
					path := c.Path()
					for _, skip := range suffixSkips {
						if ok := strings.HasSuffix(path, skip); ok {
							return true
						}
					}
					return false
				},
			}))
			new(dashboard.DashBoardController).Handler(dashboardRouter)
		}

		if data, err := json.MarshalIndent(router.Routes(), "", "  "); err == nil {
			os.WriteFile("routes.json", data, 0644)
		}
	}, func(err error) {
		// e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
		zap.L().With(zap.Int("port", port)).With(zap.Error(err)).Error("startServeAPI failed")
		cancel()
	},
		func() {
			cancel()
		})
}

// func startServeAPI(port int, handler func(router *gin.Engine), onErr func(err error), onDone func()) {
// 	trustOrigin := make(map[string]struct{})
// 	for _, s := range config.GetConfig().TrustOrigin {
// 		trustOrigin[s] = struct{}{}
// 	}
// 	zap.L().With(zap.Strings("origins", config.GetConfig().TrustOrigin)).Info("trust origin")
// 	router := gin.Default()
// 	router.Use(cors.New(cors.Config{
// 		// AllowOrigins:     []string{"https://*on"},
// 		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions, http.MethodDelete},
// 		AllowHeaders:     []string{"Content-Type", "Accept", "user-agent", "referer", "Cookie", "Authorize"},
// 		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
// 		AllowCredentials: true,
// 		AllowOriginFunc: func(origin string) bool {
// 			if _, exist := trustOrigin[origin]; exist {
// 				return true
// 			}
// 			// return exist
// 			zap.L().With(zap.String("origin", origin)).Error("reject origin")
// 			return false
// 		},
// 		MaxAge: 24 * time.Hour,
// 	}))
// 	router.Use(func(c *gin.Context) {
// 		if !shouldCompress(c.Request) {
// 			return
// 		}
// 		c.Header("Content-Encoding", "br")
// 		c.Header("Vary", "Accept-Encoding")
// 		brWriter := brotli.NewWriterV2(c.Writer, brotli.DefaultCompression)
// 		x := &CompressMidle{c.Writer, brWriter, 0}
// 		c.Writer = x
// 		defer func() {
// 			brWriter.Close()
// 			c.Header("Content-Length", strconv.Itoa(x.Length))
// 		}()
// 		c.Next()
// 	})
// 	if gin.Mode() == gin.ReleaseMode {
// 		router.SetTrustedProxies(nil)
// 	}
// 	handler(router)
// 	zap.L().With(zap.Int("port", port)).Info("start server")
// 	err := router.Run(fmt.Sprintf(":%d", port))
// 	if err != nil && onErr != nil {
// 		onErr(err)
// 		return
// 	}
// 	onDone()
// }

func startEchoServeAPI(port int, handler func(router *echo.Echo), onErr func(err error), onDone func()) {
	trustOrigin := make(map[string]struct{})
	for _, s := range config.GetConfig().TrustOrigin {
		trustOrigin[s] = struct{}{}
	}
	zap.L().With(zap.Strings("origins", config.GetConfig().TrustOrigin)).Info("trust origin")
	router := echo.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	router.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogLatency:  true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogRemoteIP: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo,
					"REQUEST",
					slog.String("latency", v.Latency.String()),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("ip", v.RemoteIP),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("latency", v.Latency.String()),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("ip", v.RemoteIP),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))
	router.Use(middleware.Recover())
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// AllowOrigins: []string{"https://labstack.com", "https://labstack.net"},
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions, http.MethodDelete},
		AllowHeaders:  []string{"Content-Type", "Accept", "user-agent", "referer", "Cookie", "Authorize", "Authorization"},
		ExposeHeaders: []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowOriginFunc: func(origin string) (bool, error) {
			if _, exist := trustOrigin[origin]; exist {
				return true, nil
			}
			// return exist
			zap.L().With(zap.String("origin", origin)).Error("reject origin")
			return false, nil
		},
	}))

	router.Use(middleware.Decompress())
	router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))
	handler(router)
	zap.L().With(zap.Int("port", port)).Info("start server")
	err := router.Start(fmt.Sprintf(":%d", port))
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

// type CompressMidle struct {
// 	gin.ResponseWriter
// 	writer *matchfinder.Writer
// 	Length int
// }

// func (g *CompressMidle) WriteString(s string) (int, error) {
// 	g.Length += len(s)
// 	return g.writer.Write([]byte(s))
// }

// func (g *CompressMidle) Write(data []byte) (int, error) {
// 	g.Length += len(data)
// 	return g.writer.Write(data)
// }

// func shouldCompress(req *http.Request) bool {
// 	if !strings.Contains(req.Header.Get("Accept-Encoding"), "br") {
// 		return false
// 	}
// 	if strings.Contains(req.URL.Path, "be/s3") {
// 		return false
// 	}
// 	if strings.Contains(req.URL.Path, "be/data/storage") {
// 		return false
// 	}
// 	if strings.Contains(req.URL.Path, "api/storage") {
// 		return false
// 	}

// 	extension := filepath.Ext(req.URL.Path)
// 	if len(extension) < 4 { // fast path
// 		return true
// 	}

// 	switch extension {
// 	case ".png", ".gif", ".jpeg", ".jpg", ".webp":
// 		return false
// 	default:
// 		return true
// 	}
// }
