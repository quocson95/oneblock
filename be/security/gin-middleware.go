package security

import (
	"be/common"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func TokenAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorize")
		// if err != nil {
		// 	zap.L().With(zap.Error(err)).Error("session cookie empty")
		// 	c.AbortWithStatus(http.StatusUnauthorized)
		// 	return
		// }

		user, err := VerifyToken(token)
		if err != nil {
			zap.L().With(zap.Error(err)).With(zap.String("token", token)).Error("verify token failed")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if cacheUser, ok := common.CacheUser.Get(token); ok {
			user := &common.User{
				UserName: cacheUser.UserName,
				Email:    cacheUser.Email,
			}
			c.Set("user", user)
			c.Next()
		}
		err = user.Read()
		if err != nil {
			zap.L().With(zap.Error(err)).Error("user not found")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("user", user)
		common.CacheUser.Add(token, *user)
		c.Next()
	}
}
