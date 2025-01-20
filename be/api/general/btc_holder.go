package api

import (
	"be/common"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func BtcHolder(c echo.Context) error {
	v := &common.HolderBtc{}
	err := v.Load(time.Now().Add(-10 * 365 * 24 * time.Hour))
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("load data error"))
	}
	return c.JSON(http.StatusOK, v)
}
