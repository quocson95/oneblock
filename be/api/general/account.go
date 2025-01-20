package api

import (
	"be/security"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AccountApi struct {
}

func (o *AccountApi) Handler(r *echo.Group) {
	r.POST("/token", o.AccountToken)
}
func (o *AccountApi) AccountToken(c echo.Context) error {
	tokenResp := security.TokenResponse{}
	data, _ := io.ReadAll(c.Request().Body)
	tokenResp.Token = string(data)
	return c.JSON(http.StatusOK, tokenResp)
}
