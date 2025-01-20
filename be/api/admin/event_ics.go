package apiadmin

import (
	"be/common"
	"be/job"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type ICSAdminController struct{}

func (i *ICSAdminController) Handler(g *echo.Group) {
	g.GET("", i.List)
	g.GET("/", i.List)
	g.GET("/log", i.ListLogCraw)
	g.GET("/:id", i.Get)
	g.POST("/", i.Add)
	g.PUT("/", i.Edit)
	g.POST("/update-cal-invest", i.UpdateCrawInvest)
}

func (i *ICSAdminController) List(c echo.Context) error {
	start, end := common.GetWeekRange(time.Now())
	ml, err := common.GetICS(start, end, 0, 1000)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get list ics failed")
		return c.NoContent(http.StatusBadRequest)

	}
	return c.JSON(http.StatusOK, ml)
}

func (i *ICSAdminController) Get(c echo.Context) error {
	id := c.Param("id")
	idInt, _ := strconv.Atoi(id)
	if idInt <= 0 {
		return c.NoContent(http.StatusOK)

	}
	ics := &common.ICS{}
	err := ics.GetById(uint(idInt))
	if err != nil {
		zap.L().With(zap.Int("id", idInt)).With(zap.Error(err)).Error("get ics failed")
		return c.NoContent(http.StatusBadRequest)

	}
	return c.JSON(http.StatusOK, ics)
}

func (i *ICSAdminController) ListLogCraw(c echo.Context) error {
	ml, err := common.GetCrawlLogs(0, 1000)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get log craw failed")
		return c.NoContent(http.StatusBadRequest)

	}
	return c.JSON(http.StatusOK, ml)
}

func (i *ICSAdminController) Add(c echo.Context) error {
	ics := &common.ICS{}
	err := c.Bind(ics)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("bind body failed")
		return c.NoContent(http.StatusBadRequest)

	}
	ics.Insert()
	return c.NoContent(http.StatusOK)
}

func (i *ICSAdminController) Edit(c echo.Context) error {
	ics := &common.ICS{}
	err := c.Bind(ics)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("bind body failed")
		return c.NoContent(http.StatusBadRequest)
	}
	icsDB := common.ICS{}
	err = icsDB.GetById(ics.ID)
	if err != nil {
		zap.L().With(zap.Uint("id", ics.ID)).With(zap.Error(err)).Error("patch ics failed")
		return c.NoContent(http.StatusBadRequest)
	}
	err = icsDB.Updates(map[string]interface{}{"start_unix": ics.StartUnix, "end_unix": ics.EndUnix, "name": ics.Name, "desp": ics.Desp})
	if err != nil {
		zap.L().With(zap.Uint("id", ics.ID)).With(zap.Error(err)).Error("patch ics failed")
		return c.NoContent(http.StatusBadRequest)
	}
	icsDB.GetById(ics.ID)
	return c.JSON(http.StatusOK, icsDB)
}

func (i *ICSAdminController) UpdateCrawInvest(c echo.Context) error {
	events, err := job.ParseCrawlInvestCal(c.Request().Body)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("craw failed")
		return c.NoContent(http.StatusBadRequest)
	}
	job.InsertEventInvestCal(events)
	return c.JSON(http.StatusOK, events)
}
