package api

import (
	"be/common"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func BtcHolder(c *gin.Context) {
	v := &common.HolderBtc{}
	err := v.Load(time.Now().Add(-10 * 365 * 24 * time.Hour))
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		c.AbortWithError(http.StatusBadRequest, errors.New("load data error"))
	}
	c.JSON(http.StatusOK, v)
}
