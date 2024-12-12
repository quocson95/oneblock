package api

import (
	"be/common"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Define the struct for the JSON data

func BtcGold(c *gin.Context) {
	data, err := os.ReadFile("raw_data/btc_gold.json")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, "read file error")
	}
	correlationData := common.CorrelationData{}
	json.Unmarshal(data, &correlationData)
	c.JSON(http.StatusOK, correlationData)
}

func BtcGoldAgressApi(c *gin.Context) {
	correlationData := common.CorrelationData{}
	err := common.Load(common.BtcGold, func(data []byte) error {
		return json.Unmarshal(data, &correlationData)
	})
	if err != nil {
		zap.L().With(zap.Error(err)).Error("load data error")
		c.AbortWithStatusJSON(http.StatusBadRequest, "load data error")
	}

	t := time.Now()
	fromTime := time.Date(t.Year()-6, t.Month(), 0, 0, 0, 0, 0, t.Location())
	m := &common.BtcGoldAgress{}
	m.Agresss(correlationData, fromTime)
	c.JSON(http.StatusOK, m)
}
