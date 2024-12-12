package api

import (
	"be/common"
	"be/craw"
	"net/http"

	"github.com/gin-gonic/gin"
)

func BtcEthStatic(e *gin.Context) {
	btcPrice, err := craw.BtcPriceSummary()
	if err != nil {
		e.JSON(http.StatusOK, "")
		return
	}
	ethGas, err := craw.EthGas()
	if err != nil {
		e.JSON(http.StatusOK, "")
		return
	}
	btcEth := common.BtcEth{
		BtcCurrentPrice: *btcPrice,
		EthGas:          *ethGas,
	}
	e.JSON(http.StatusOK, btcEth)
	return
}
