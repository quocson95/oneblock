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
		e.JSON(http.StatusOK, "")
		return nil
	}
	ethGas, err := craw.EthGas()
	if err != nil {
		e.JSON(http.StatusOK, "")
		return nil
	}
	btcEth := common.BtcEth{
		BtcCurrentPrice: *btcPrice,
		EthGas:          *ethGas,
	}
	e.JSON(http.StatusOK, btcEth)
	return nil
}
