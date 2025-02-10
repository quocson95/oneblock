package dashboard

import (
	api "be/api/general"
	"be/common"
	"be/config"
	"be/security"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type DashBoardController struct{}

func (d *DashBoardController) Handler(r *echo.Group) {
	{
		googleSSOHandler := api.NewOath2Api(config.GetConfig().GoogleConsoleCustomer, security.SecretJwtAuthDashboard, func(userInfo map[string]interface{}) (*common.User, error) {
			email, _ := userInfo["email"].(string)
			fmt.Println(userInfo)
			// map[email:dangquocson1995@gmail.com family_name:Dang given_name:X id:109085060275590059569 name:X Dang
			//  picture:https://lh3.googleusercontent.com/a/ACg8ocK7SpiMGJQSDewLBZ27jys8XHQULLRhYBuInq1xiPdHxX6YFnZR=s96-c verified_email:true]
			user := &common.User{}
			err := user.FindByEmail(email, []common.RoleUser{common.RoleUserAdmin, common.RoluserCustomer})
			if err == gorm.ErrRecordNotFound {
				u := &common.User{
					Email:         email,
					Role:          common.RoluserCustomer,
					LastLoginUnix: time.Now().Unix(),
				}
				u.Picture, _ = userInfo["picture"].(string)
				u.UserName, _ = userInfo["name"].(string)
				u.Create()
				user = u
			}
			if err == nil {
				picture, _ := userInfo["picture"].(string)
				changes := make(map[string]interface{})
				if user.Picture != picture {
					changes["picture"] = picture
				}
				userName, _ := userInfo["name"].(string)
				if user.UserName != userName {
					changes["user_name"] = userName
				}
				changes["last_login_unix"] = time.Now().Unix()
				user.Update(changes)
			}
			return user, err

		}, nil)
		googleSSOHandler.Handler(r.Group("/sso/google"))
	}
	// user
	{
		new(UserController).Handler(r.Group("/user"))
	}
	// payment
	{
		// new(PaymentController).Handler(r.Group("/payment"))
	}
	// Payos
	{
		cfg := config.GetConfig().PayOS
		NewPayOS(cfg.ClientID, cfg.ApiKey, cfg.ChecksumKey).Handler(r.Group("/payment/payos"))
	}
	// Manager User
	{
		gCustomer := r.Group("/customer")
		gCustomer.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				user := security.GetUserCtx(c)
				if user == nil {
					return echo.NewHTTPError(http.StatusBadRequest, "invalid user")
				}
				if user.Role != common.RoleUserAdmin && user.Role != common.RoleUserManager {
					return echo.NewHTTPError(http.StatusBadRequest, "unauthorization user")
				}
				return next(c)
			}
		})
		new(Customer).Handler(gCustomer)
		new(CopyTradeOrder).Handler(gCustomer.Group("/copy-trade"))

	}
}
