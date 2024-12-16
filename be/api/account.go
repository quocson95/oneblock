package api

import (
	"be/security"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AccountApi struct {
}

func (o *AccountApi) Handler(r gin.IRoutes) {
	r.POST("/token", o.AccountToken)
}
func (o *AccountApi) AccountToken(c *gin.Context) {
	tokenResp := security.TokenResponse{}
	data, _ := io.ReadAll(c.Request.Body)
	tokenResp.Token = string(data)
	c.JSON(http.StatusOK, tokenResp)
}
