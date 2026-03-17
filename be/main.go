package main

import (
	apiadmin "be/api/admin"
	apicuuho "be/api/cuuho"
	"be/api/dashboard"
	api "be/api/general"
	"be/bot"
	"be/common"
	"be/config"
	"be/database"
	"be/security"
	"be/sp500"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/andybalholm/brotli"
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
	if config.GetConfig().Idrivee2Endpoint != "" {
		common.Idrivee2S3Helper.Init(config.GetConfig().Idrivee2Endpoint, config.GetConfig().Idrivee2Region, config.GetConfig().Idrivee2AccessKey, config.GetConfig().Idrivee2SecretKey)
	}
	port := config.GetConfig().Port
	if port <= 0 {
		port = 80
	}

	if err := database.InitDB(config.GetConfig().PostgressDsn); err != nil {
		panic(fmt.Errorf("init db failed: %s", err.Error()))
	}
	database.DB.AutoMigrate(new(common.Mdx), new(common.User), new(common.ICS), new(common.CrawlLog), new(common.S3ObjectSync),
		new(common.Payment), new(common.Plan), new(common.Subscribe), new(common.CopyTradeOrder))
	common.GetDB = func() *gorm.DB {
		return database.DB
	}
	go createPlan()
	// common.InitRedis(config.GetConfig().Redis.Addr, )
	createDefaultUser()
	go bot.InitTeleBot(config.GetConfig().TeleBot.Token, config.GetConfig().TeleBot.ChatId, config.GetConfig().TeleBot.ChannelUsername)

	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx
	// job.JobCrawlInvestingCalendar()
	// go job.StartJob(ctx)
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
		new(apicuuho.CuuhoController).Handler(beRouter.Group("/cuuho"))
		new(api.NuoiToiController).Handler(beRouter.Group("/nuoitoi"))
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
		api.NewOath2Api(config.GetConfig().GoogleConsole, security.SecretJwtAuth, nil, nil).Handler(beRouter.Group("/auth"))
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
				TokenLookup: "header:Authorization,header:Authorization:Bearer ",
				// ...
			}))
			new(apiadmin.MdxAdminController).Handler(adminRouter.Group("/mdx"))
			new(apiadmin.ICSAdminController).Handler(adminRouter.Group("/ics"))
		}
		// dashboard
		{
			// be/dashboard/sso/google
			suffixSkips := []string{"/sso/google", "/sso/google/callback", "/payment/payos/webhook"}
			suffixContainSkips := []string{"/payment/payos/qrcode/"}
			dashboardRouter := beRouter.Group("/dashboard")
			dashboardRouter.Use(echojwt.WithConfig(echojwt.Config{
				// ...
				SigningKey:     []byte(security.SecretJwtAuthDashboard),
				SuccessHandler: security.SuccessHandlerDashboardUser(database.DB),
				// ContinueOnIgnoredError: false,
				ErrorHandler: func(c echo.Context, err error) error {
					zap.L().With(zap.String("token", c.Request().Header.Get("Authorization"))).With(zap.Error(err)).Error("jwt handler error")
					return c.NoContent(http.StatusUnauthorized)
				},
				TokenLookup: "header:Authorization,header:Authorization:Bearer ,cookie:authorization",
				Skipper: func(c echo.Context) bool {
					path := c.Path()
					for _, skip := range suffixSkips {
						if ok := strings.HasSuffix(path, skip); ok {
							return true
						}
					}
					for _, skip := range suffixContainSkips {
						if ok := strings.Contains(path, skip); ok {
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

func startEchoServeAPI(port int, handler func(router *echo.Echo), onErr func(err error), onDone func()) {
	trustOrigin := make(map[string]struct{})
	for _, s := range config.GetConfig().TrustOrigin {
		trustOrigin[s] = struct{}{}
	}
	zap.L().With(zap.Strings("origins", config.GetConfig().TrustOrigin)).Info("trust origin")
	router := echo.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	router.Use(middlewareLog(logger))
	router.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(200))))
	router.Use(middleware.Recover())
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		// AllowOrigins: []string{"https://labstack.com", "https://labstack.net"},
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodOptions, http.MethodDelete},
		AllowHeaders:  []string{"Content-Type", "Accept", "user-agent", "referer", "Cookie", "Authorize", "Authorization"},
		ExposeHeaders: []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowOriginFunc: func(origin string) (bool, error) {
			// if _, exist := trustOrigin[origin]; exist {
			// 	return true, nil
			// }
			// // return exist
			// zap.L().With(zap.String("origin", origin)).Error("reject origin")
			// return false, nil
			return true, nil
		},
	}))

	router.Use(middleware.Decompress())
	// router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
	// 	Level: 5,
	// 	Skipper: func(c echo.Context) bool {
	// 		return strings.HasPrefix(c.Response().Header().Get("Content-Type"), "image/")
	// 	},
	// }))
	router.Use(common.BrotliWithConfig(common.BrotliConfig{
		Level: brotli.DefaultCompression,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Response().Header().Get("Content-Type"), "image/")
		},
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
	// emails := []string{
	// 	"dangquocson1995@gmail.com",
	// 	"nghuuloc512@gmail.com",
	// 	"nguyentrungbmt17@gmail.com",
	// 	"haotran1689@gmail.com",
	// 	"dtoan.bui@gmail.com",
	// }
	// for _, email := range emails {
	// 	u := common.User{
	// 		Email:        email,
	// 		Role:         common.RoleUserAdmin,
	// 		UsdtInWallet: 0,
	// 	}
	// 	if err := u.Create(); err != nil {
	// 		zap.L().With(zap.Error(err)).With(zap.String("email", email)).Error("add user failed")
	// 	}

	// }
	// u:=common.User{
	// 	Email: "sondq.1024@gmail.com",
	// }
}

func startMetricsHandler(port int) {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func createPlan() {
	plan := &common.Plan{
		Id:             1,
		Name:           "Plan 1",
		Price:          100,
		Currency:       "$",
		Desp:           "This is plan 1",
		PlanRenewType:  common.PlanRenewTypeMonth,
		DurationExtend: "86400s",
	}
	plan.PriceDisp = fmt.Sprintf("%d %s", plan.Price, plan.Currency)
	plan.Create()
}

func middlewareLog(logger *slog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogLatency:  true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogRemoteIP: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []slog.Attr{slog.String("latency", v.Latency.String()),
				slog.String("uri", v.URI),
				slog.Int("status", v.Status),
				slog.String("ip", v.RemoteIP),
				slog.String("content-encoding", strings.Join(v.Headers["content-encoding"], ","))}
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo,
					"REQUEST",
					attrs...,
				)
			} else {
				attrs = append(attrs, slog.String("err", v.Error.Error()))
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					attrs...,
				)
			}
			return nil
		},
	})
}
