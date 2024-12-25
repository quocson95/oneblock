package apiadmin

import (
	"be/common"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ICSAdminController struct{}

func (i *ICSAdminController) Handler(g gin.IRoutes) {
	g.GET("", i.List)
	g.GET("/", i.List)
	g.GET("/log", i.ListLogCraw)
	g.GET("/:id", i.Get)
	g.POST("/", i.Add)
	g.PUT("/", i.Edit)
}

func (i *ICSAdminController) List(c *gin.Context) {
	start, end := common.GetWeekRange(time.Now())
	ml, err := common.GetICS(start, end, 0, 1000)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get list ics failed")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, ml)
}

func (i *ICSAdminController) Get(c *gin.Context) {
	id := c.Param("id")
	idInt, _ := strconv.Atoi(id)
	if idInt <= 0 {
		c.Abort()
		return
	}
	ics := &common.ICS{}
	err := ics.GetById(uint(idInt))
	if err != nil {
		zap.L().With(zap.Int("id", idInt)).With(zap.Error(err)).Error("get ics failed")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, ics)
}

func (i *ICSAdminController) ListLogCraw(c *gin.Context) {
	ml, err := common.GetCrawlLogs(0, 1000)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get log craw failed")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusOK, ml)
}

func (i *ICSAdminController) Add(c *gin.Context) {
	ics := &common.ICS{}
	err := c.ShouldBindBodyWithJSON(ics)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("bind body failed")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ics.Insert()
}

func (i *ICSAdminController) Edit(c *gin.Context) {
	ics := &common.ICS{}
	err := c.ShouldBindBodyWithJSON(ics)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("bind body failed")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	icsDB := common.ICS{}
	err = icsDB.GetById(ics.ID)
	if err != nil {
		zap.L().With(zap.Uint("id", ics.ID)).With(zap.Error(err)).Error("patch ics failed")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	err = icsDB.Updates(map[string]interface{}{"start_unix": ics.StartUnix, "end_unix": ics.EndUnix, "name": ics.Name, "desp": ics.Desp})
	if err != nil {
		zap.L().With(zap.Uint("id", ics.ID)).With(zap.Error(err)).Error("patch ics failed")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	icsDB.GetById(ics.ID)
	c.JSON(http.StatusOK, icsDB)
}
