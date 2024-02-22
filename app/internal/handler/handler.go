package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gamepkw/google-oauth2/app/internal/helpers/pages"
	"github.com/gamepkw/google-oauth2/app/internal/model"
	oAuthService "github.com/gamepkw/google-oauth2/app/internal/service"
	"github.com/gamepkw/google-oauth2/tools/logger"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type OAuthHandler struct {
	oAuthService      oAuthService.OAuthService
	googleOauthConf   *oauth2.Config
	facebookOauthConf *oauth2.Config
	githubOauthConf   *oauth2.Config
}

func NewOAuthHandler(e *echo.Echo, auth oAuthService.OAuthService, googleOauthConf, facebookOauthConf, githubOauthConf *oauth2.Config) {
	handler := &OAuthHandler{
		oAuthService:      auth,
		googleOauthConf:   googleOauthConf,
		facebookOauthConf: facebookOauthConf,
		githubOauthConf:   githubOauthConf,
	}

	//google
	e.GET("/login-google", handler.LoginGoogle)
	e.GET("/login-google-implicit", handler.LoginGoogleImplicit)
	e.POST("/revoke-token-google", handler.RevokeTokenGoogle)
	e.POST("/callback-google", handler.CallbackFromGoogle)
	e.POST("/get-new-token", handler.GetNewAccessToken)

	//facebook
	e.GET("/login-facebook", handler.LoginFacebook)

	//github
	e.GET("/login-github", handler.LoginGithub)
	e.POST("/callback-github", handler.CallbackFromGithub)
}

var oauthStateStringGl = viper.GetString("oauthStateString")

func (a *OAuthHandler) Main(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)

	return c.HTMLBlob(http.StatusOK, []byte(pages.IndexPage))
}

func (a *OAuthHandler) LoginGoogle(c echo.Context) error {
	ctx := c.Request().Context()

	oauthStateString := viper.GetString("oauth_state")

	authURL, err := a.oAuthService.GetGoogleAuthUrl(ctx, oauthStateString)
	if err != nil {
		return err
	}

	fmt.Println(authURL)

	return c.JSON(http.StatusOK, authURL)
}

func (a *OAuthHandler) RevokeTokenGoogle(c echo.Context) error {
	ctx := c.Request().Context()

	var request model.RevokeAccessTokenRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	if err := a.oAuthService.RevokeAccessToken(ctx, request.AccessToken); err != nil {
		return c.JSON(http.StatusBadRequest, nil)
	}

	return c.JSON(http.StatusOK, nil)
}

func (a *OAuthHandler) LoginGoogleImplicit(c echo.Context) error {
	ctx := c.Request().Context()

	oauthStateString := viper.GetString("oauth_state")

	authURL, err := a.oAuthService.GetGoogleAuthImplicitUrl(ctx, oauthStateString)
	if err != nil {
		return err
	}

	fmt.Println(authURL)

	return c.JSON(http.StatusOK, authURL)
}

func (a *OAuthHandler) GetNewAccessToken(c echo.Context) error {
	ctx := c.Request().Context()

	var request model.GetNewAccessTokenRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	resp := a.oAuthService.GetNewAccessToken(ctx, request.RefreshToken)

	fmt.Println(resp)

	return c.JSON(http.StatusOK, resp)
}

func (a *OAuthHandler) CallbackFromGoogle(c echo.Context) error {
	logger.Log.Info("Callback-google..")

	var request model.CallbackGoogleRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	state := c.FormValue("state")
	logger.Log.Info(state)
	if state != oauthStateStringGl {
		logger.Log.Info("invalid oauth state, expected " + oauthStateStringGl + ", got " + state + "\n")
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	token, err := a.oAuthService.ExchangeAccessTokenGoogle(context.Background(), request)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.CallbackGoogleResponse{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken})
}

func (a *OAuthHandler) CallbackFromGithub(c echo.Context) error {
	logger.Log.Info("Callback-github..")

	var request model.CallbackGoogleRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, err.Error())
	}

	state := c.FormValue("state")
	logger.Log.Info(state)
	if state != oauthStateStringGl {
		logger.Log.Info("invalid oauth state, expected " + oauthStateStringGl + ", got " + state + "\n")
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	token, err := a.oAuthService.ExchangeAccessTokenGithub(context.Background(), request)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, model.CallbackGoogleResponse{AccessToken: token.AccessToken})
}

func (a *OAuthHandler) LoginFacebook(c echo.Context) error {
	ctx := c.Request().Context()

	var oauthStateStringGl = viper.GetString("oauthStateString")

	authURL := a.oAuthService.GetFacebookAuthUrl(ctx, oauthStateStringGl)

	fmt.Println(authURL)

	return c.JSON(http.StatusOK, authURL)
}

func (a *OAuthHandler) LoginGithub(c echo.Context) error {
	ctx := c.Request().Context()

	var oauthStateStringGl = viper.GetString("oauthStateString")

	authURL := a.oAuthService.GetGithubAuthUrl(ctx, oauthStateStringGl)

	fmt.Println(authURL)

	return c.JSON(http.StatusOK, authURL)
}
