package dashboard

import (
	"be/common"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type UserController struct{}

func (r *UserController) Handler(g *echo.Group) {
	g.GET("", r.GetUser)
	g.GET("/", r.GetUser)
}

func (r *UserController) GetUser(c echo.Context) error {
	u, ok := c.Get("user").(*common.User)
	if !ok || len(u.Email) == 0 {
		zap.L().Error("user email is empty")
		return c.NoContent(http.StatusBadRequest)
	}
	user, err := common.GetUserInCache(u.Email)
	if err != nil {
		zap.L().With(zap.String("email", u.Email)).With(zap.Error(err)).Error("get user failed")
		return c.NoContent(http.StatusBadRequest)
	}
	return c.JSON(http.StatusOK, user)
}
