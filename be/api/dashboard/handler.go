package dashboard

import (
	api "be/api/general"
	"be/common"
	"be/config"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type DashBoardController struct{}

func (d *DashBoardController) Handler(r *echo.Group) {
	{

		googleSSOHandler := api.NewOath2Api(config.GetConfig().GoogleConsoleCustomer, func(userInfo map[string]interface{}) (*common.User, error) {
			email, _ := userInfo["email"].(string)
			// fmt.Println(userInfo)
			// map[email:dangquocson1995@gmail.com family_name:Dang given_name:X id:109085060275590059569 name:X Dang
			//  picture:https://lh3.googleusercontent.com/a/ACg8ocK7SpiMGJQSDewLBZ27jys8XHQULLRhYBuInq1xiPdHxX6YFnZR=s96-c verified_email:true]
			user := &common.User{}
			err := user.FindByEmail(email, common.RoleUserAdmin, common.RoluserMember)
			if err == gorm.ErrRecordNotFound {
				u := &common.User{
					Email: email,
				}
				u.Picture, _ = userInfo["picture"].(string)
				u.UserName, _ = userInfo["name"].(string)
				u.Create()
				user = u
			}
			if err == nil {
				picture, _ := userInfo["picture"].(string)
				if user.Picture != picture {
					user.Update(map[string]interface{}{"picture": picture})
				}
			}
			return user, err

		}, nil)
		googleSSOHandler.Handler(r.Group("/sso/google"))
	}
}
