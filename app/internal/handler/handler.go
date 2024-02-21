package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

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

	code := request.AuthCode
	logger.Log.Info(code)

	fmt.Println("code", code)

	if code == "" {
		logger.Log.Warn("Code not found..")
		return c.String(http.StatusOK, "Code Not Found to provide AccessToken..\n")
	}

	opts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
	}

	token, err := a.googleOauthConf.Exchange(context.Background(), code, opts...)
	if err != nil {
		logger.Log.Error("oauthConfGl.Exchange() failed with " + err.Error() + "\n")
		return err
	}

	fmt.Println("accessToken", token)

	logger.Log.Info("TOKEN>> AccessToken>> " + token.AccessToken)
	logger.Log.Info("TOKEN>> Expiration Time>> " + token.Expiry.String())
	logger.Log.Info("TOKEN>> RefreshToken>> " + token.RefreshToken)

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	if err != nil {
		logger.Log.Error("Get: " + err.Error() + "\n")
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("ReadAll: " + err.Error() + "\n")
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	logger.Log.Info("parseResponseBody: " + string(response) + "\n")
	fmt.Println(string(response))
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

	code := request.AuthCode
	logger.Log.Info(code)

	fmt.Println("code", code)

	if code == "" {
		logger.Log.Warn("Code not found..")
		return c.String(http.StatusOK, "Code Not Found to provide AccessToken..\n")
	}

	token, err := a.githubOauthConf.Exchange(context.Background(), code)
	if err != nil {
		logger.Log.Error("oauthConfGl.Exchange() failed with " + err.Error() + "\n")
		return err
	}

	fmt.Println("accessToken", token)

	logger.Log.Info("TOKEN>> AccessToken>> " + token.AccessToken)
	logger.Log.Info("TOKEN>> Expiration Time>> " + token.Expiry.String())
	logger.Log.Info("TOKEN>> RefreshToken>> " + token.RefreshToken)

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	if err != nil {
		logger.Log.Error("Get: " + err.Error() + "\n")
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("ReadAll: " + err.Error() + "\n")
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	logger.Log.Info("parseResponseBody: " + string(response) + "\n")
	fmt.Println(string(response))
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
