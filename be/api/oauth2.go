package api

import (
	"be/common"
	"be/config"
	"be/security"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Oauth2Api struct {
	oauth2Config *oauth2.Config
	oauth2State  string
	callbackSSo  string
}

func NewOath2Api(googleCfg config.GoogleConsole) *Oauth2Api {
	// Set up OAuth2 configuration
	oauth2Config := &oauth2.Config{
		ClientID:     googleCfg.ID,
		ClientSecret: googleCfg.Secret,
		RedirectURL:  googleCfg.CallbackSSO,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	o := &Oauth2Api{
		oauth2Config: oauth2Config,
		callbackSSo:  googleCfg.CallbackSSO,
		oauth2State:  common.Sha265Random(),
	}
	return o
}

func (o *Oauth2Api) Handler(r *echo.Group) {
	r.GET("/google_callback", o.GooleOauth2Callback)
	r.GET("/google_signin", o.GoogleSignIn)
	r.POST("/login", o.Login)

}

func (o *Oauth2Api) GoogleSignIn(c echo.Context) error {
	urlGoogleLogin := fmt.Sprintf(`https://accounts.google.com/o/oauth2/v2/auth?scope=openid email profile&access_type=offline&include_granted_scopes=true&response_type=code&state=%s&redirect_uri=%s&client_id=%s`,
		o.oauth2State, o.callbackSSo, o.oauth2Config.ClientID)
	return c.Redirect(http.StatusFound, urlGoogleLogin)
}

func (o *Oauth2Api) GooleOauth2Callback(c echo.Context) error {
	state := c.QueryParam("state")
	if state != o.oauth2State {
		// return c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid state")
	}

	code := c.QueryParam("code")
	if code == "" {
		// return c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
		return echo.NewHTTPError(http.StatusBadRequest, "Missing code")
	}

	token, err := o.oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		// return c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	client := o.oauth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		// return c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	defer resp.Body.Close()

	// Process user info
	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		// return c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	email := userInfo["email"].(string)
	user := &common.User{}
	err = user.FindByEmail(email)
	// reqUrl := c.Request.Proto
	redirectUrl := fmt.Sprintf("%s?id=%s&errCode=%d&errStr=%s", config.GetConfig().GoogleConsole.RedirectURI, "", 100, "user-not-found")
	if err != nil {
		zap.L().With(zap.Error(err)).With(zap.String("email", email)).Error("find user failed")
		// c.AbortWithError(http.StatusBadRequest, errors.New("user-not-found"))
		return c.Redirect(http.StatusFound, redirectUrl)
	}
	user.Update(map[string]interface{}{"last_login": time.Now()})
	tokenResp, err := security.CreateToken(user)
	if err != nil {
		zap.L().With(zap.String("email", user.Email)).With(zap.Error(err)).Error("create token failed")
		// c.AbortWithError(http.StatusBadRequest, errors.New("create token failed"))
		return c.Redirect(http.StatusFound, redirectUrl)
	}
	zap.L().With(zap.String("user", user.Email)).Info("redirect user ok")
	redirectUrl = fmt.Sprintf("%s?id=%s&errCode=%d&errStr=%s", config.GetConfig().GoogleConsole.RedirectURI, tokenResp.Token, 0, "")
	// c.SetCookie("token", tokenResp.Token, 86400, "", "https://editor.oneblock.vn", false, false)
	return c.Redirect(http.StatusFound, redirectUrl)
	// c.JSON(http.StatusOK, tokenResp)
}

func (o *Oauth2Api) Login(c echo.Context) error {
	tokenResp := security.TokenResponse{}
	data, _ := io.ReadAll(c.Request().Body)
	tokenResp.Token = string(data)
	return c.JSON(http.StatusOK, tokenResp)
}
