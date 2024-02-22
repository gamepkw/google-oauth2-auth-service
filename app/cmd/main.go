package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gamepkw/google-oauth2/app/internal/config"
	_oauthHandler "github.com/gamepkw/google-oauth2/app/internal/handler"
	_oauthService "github.com/gamepkw/google-oauth2/app/internal/service"
	"github.com/gamepkw/google-oauth2/tools/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	facebookOAuth "golang.org/x/oauth2/facebook"
	githubOAuth "golang.org/x/oauth2/github"
	googleOAuth "golang.org/x/oauth2/google"
)

func main() {
	env := os.Getenv("DOCKER_ENV")
	config.InitializeViper()
	logger.InitializeZapCustomLogger()

	googleOauthConf := &oauth2.Config{
		ClientID:     viper.GetString("google.clientID"),
		ClientSecret: viper.GetString("google.clientSecret"),
		RedirectURL:  viper.GetString("google.redirectURL"),
		Scopes: []string{
			viper.GetString("google.scopes-profile"),
			viper.GetString("google.scopes-email"),
			viper.GetString("google.scopes-openid"),
			viper.GetString("google.scopes-calendar"),
		},
		Endpoint: googleOAuth.Endpoint,
	}

	facebookOauthConf := &oauth2.Config{
		ClientID:     viper.GetString("facebook.appID"),
		ClientSecret: viper.GetString("facebook.appSecret"),
		RedirectURL:  viper.GetString("facebook.redirectURL"),
		Endpoint:     facebookOAuth.Endpoint,
		Scopes:       []string{"email"},
	}

	githubOauthConf := &oauth2.Config{
		ClientID:     viper.GetString("github.clientID"),
		ClientSecret: viper.GetString("github.clientSecret"),
		RedirectURL:  viper.GetString("github.redirectURL"),
		Endpoint:     githubOAuth.Endpoint,
		Scopes:       []string{"email"},
	}

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:3001", "http://localhost:3002"},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	oauthService := _oauthService.NewOAuthService(googleOauthConf, facebookOauthConf, githubOauthConf)
	_oauthHandler.NewOAuthHandler(e, oauthService, googleOauthConf, facebookOauthConf, githubOauthConf)

	if env == "docker" {
		logger.Log.Info("Started running on http://localhost:" + viper.GetString("port"))
		log.Fatal(e.Start(":" + viper.GetString("port")))
	} else {
		logger.Log.Info("Started running on http://localhost:8090")
		log.Fatal(e.Start(":8090"))
	}
}
