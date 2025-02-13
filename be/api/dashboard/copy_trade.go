package dashboard

import (
	"be/common"
	"be/security"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type CopyTradeOrder struct{}

func (t *CopyTradeOrder) Handler(g *echo.Group) {
	g.GET("", t.GetAllCopyTrade)
	g.GET("/", t.GetAllCopyTrade)

	g.POST("", t.Create)
	g.PUT("", t.Update)

	g.GET("/performance", t.PerformanceOverview)
	g.GET("/perf-data-chart", t.PerfChartData)
	// g.DELETE("/", t.de)
}

func (t *CopyTradeOrder) Create(c echo.Context) error {
	user := security.GetUserCtx(c)
	if user == nil {
		return c.NoContent(http.StatusBadRequest)
	}
	order := &common.CopyTradeOrder{}
	err := c.Bind(order)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("bind body failed")
		return c.NoContent(http.StatusBadRequest)
	}
	err = order.Insert()
	if err != nil {
		zap.L().With(zap.Error(err)).Error("insert copy trade order failed")
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, order)
}

func (t *CopyTradeOrder) Update(c echo.Context) error {
	user := security.GetUserCtx(c)
	if user == nil {
		return c.NoContent(http.StatusBadRequest)
	}
	order := &common.CopyTradeOrder{}
	err := c.Bind(order)
	if err != nil || order.ID <= 0 {
		zap.L().With(zap.Error(err)).Error("bind body failed")
		return c.NoContent(http.StatusBadRequest)
	}
	changes := make(map[string]interface{})
	if order.StatusOrder > common.CopyTradeOrderStatusUnknow {
		changes["status_order"] = order.StatusOrder
	}
	err = order.Updates(changes)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("insert copy trade order failed")
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, order)
}

type opts struct {
	FromUnix    int `json:"fromUnix,omitempty"`
	ToUnix      int `json:"toUnix,omitempty"`
	StatusOrder int `json:"statusOrder,omitempty"`
	Offset      int `json:"offset,omitempty"`
	Limit       int `json:"limit,omitempty"`
}

func (t *CopyTradeOrder) GetAllCopyTrade(c echo.Context) error {
	user := security.GetUserCtx(c)
	if user == nil {
		zap.L().Error("user not allow call")
		return c.NoContent(http.StatusBadRequest)
	}
	req := &opts{}
	err := c.Bind(req)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("bind body failed")
		return c.NoContent(http.StatusBadRequest)
	}
	if req.Offset < 0 {
		req.Offset = 0
	}
	if req.Limit <= 0 {
		req.Limit = 100000
	}
	orders, err := common.GetAllCopyTradeOrder(req.StatusOrder, time.Unix(int64(req.FromUnix), 0), time.Unix(int64(req.ToUnix), 0), req.Offset, req.Limit)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get all copy trade order failed")
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, orders)
}

func (t *CopyTradeOrder) PerformanceOverview(c echo.Context) error {
	orders, err := common.OverviewCopyTradeOrder()
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get overview copy trade order failed")
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, orders)
}

type ColData struct {
	Name string   `json:"name,omitempty"`
	Data []string `json:"data,omitempty"`
}
type CopyTradeOrderChartData struct {
	Xaxis []string  `json:"xaxis,omitempty"`
	Yaxis []ColData `json:"yaxis,omitempty"`
}

func (t *CopyTradeOrder) PerfChartData(c echo.Context) error {
	orders, err := common.OverviewCopyTradeOrder()
	if err != nil {
		zap.L().With(zap.Error(err)).Error("get overview copy trade order failed")
		return c.NoContent(http.StatusBadRequest)
	}
	chartData := CopyTradeOrderChartData{}
	colDataROI := ColData{Name: "ROI"}
	colDataPNL := ColData{Name: "PNL"}

	for _, order := range orders {
		x, _ := common.ISOWeekToTime(order.Year, order.Week)
		chartData.Xaxis = append(chartData.Xaxis, fmt.Sprintf("%2d/%2d", x.Month(), x.Day()))
		colDataROI.Data = append(colDataROI.Data, fmt.Sprintf("%.2f", order.Roi))
		colDataPNL.Data = append(colDataPNL.Data, strconv.Itoa(order.PNL))
	}
	chartData.Yaxis = append(chartData.Yaxis, colDataPNL)
	chartData.Yaxis = append(chartData.Yaxis, colDataROI)
	return c.JSON(http.StatusOK, chartData)
}
