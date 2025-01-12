package api

import (
	"be/common"
	"be/craw"
	"net/http"

	"github.com/labstack/echo/v4"
)

func BtcEthStatic(e echo.Context) error {
	btcPrice, err := craw.BtcPriceSummary()
	if err != nil {
		return e.JSON(http.StatusOK, "")

	}
	ethGas, err := craw.EthGas()
	if err != nil {
		return e.JSON(http.StatusOK, "")

	}
	btcEth := common.BtcEth{
		BtcCurrentPrice: *btcPrice,
		EthGas:          *ethGas,
	}
	return e.JSON(http.StatusOK, btcEth)
}
