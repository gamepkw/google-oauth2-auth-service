package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gamepkw/google-oauth2/app/internal/model"
	"github.com/gamepkw/google-oauth2/tools/logger"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type oAuthService struct {
	googleOauthConf   *oauth2.Config
	facebookOauthConf *oauth2.Config
	githubOauthConf   *oauth2.Config
}

func NewOAuthService(googleOauthConf, facebookOauthConf, githubOauthConf *oauth2.Config) OAuthService {
	return &oAuthService{
		googleOauthConf:   googleOauthConf,
		facebookOauthConf: facebookOauthConf,
		githubOauthConf:   githubOauthConf,
	}
}

type OAuthService interface {
	GetGoogleAuthUrl(ctx context.Context, oauthStateString string) (string, error)
	GetGoogleAuthImplicitUrl(ctx context.Context, oauthStateString string) (string, error)
	RevokeAccessToken(ctx context.Context, accessToken string) error
	GetNewAccessToken(ctx context.Context, refreshToken string) string
	GetFacebookAuthUrl(ctx context.Context, oauthStateString string) string
	GetGithubAuthUrl(ctx context.Context, oauthStateString string) string
	ExchangeAccessTokenGoogle(ctx context.Context, request model.CallbackGoogleRequest) (*oauth2.Token, error)
	ExchangeAccessTokenGithub(ctx context.Context, request model.CallbackGoogleRequest) (*oauth2.Token, error)
}

func (a *oAuthService) GetGoogleAuthUrl(ctx context.Context, oauthStateString string) (string, error) {

	URL, err := url.Parse(a.googleOauthConf.Endpoint.AuthURL)
	if err != nil {
		return "", err
	}

	parameters := url.Values{}
	parameters.Add("client_id", a.googleOauthConf.ClientID)
	parameters.Add("scope", strings.Join(a.googleOauthConf.Scopes, " "))
	parameters.Add("redirect_uri", a.googleOauthConf.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("access_type", "offline")
	parameters.Add("state", oauthStateString)

	URL.RawQuery = parameters.Encode()
	authURL := URL.String()

	return authURL, nil
}

func (a *oAuthService) GetGoogleAuthImplicitUrl(ctx context.Context, oauthStateString string) (string, error) {

	URL, err := url.Parse(a.googleOauthConf.Endpoint.AuthURL)
	if err != nil {
		return "", err
	}

	parameters := url.Values{}
	parameters.Add("client_id", a.googleOauthConf.ClientID)
	parameters.Add("scope", strings.Join(a.googleOauthConf.Scopes, " "))
	parameters.Add("redirect_uri", a.googleOauthConf.RedirectURL)
	parameters.Add("response_type", "token")
	parameters.Add("state", oauthStateString)

	URL.RawQuery = parameters.Encode()
	authURL := URL.String()

	return authURL, nil
}

func (a *oAuthService) GetFacebookAuthUrl(ctx context.Context, oauthStateString string) string {
	authURL := a.facebookOauthConf.AuthCodeURL(oauthStateString)
	fmt.Println(oauthStateString)
	return authURL
}
func (a *oAuthService) CallbackFromGoogle(ctx context.Context, oauthStateString string) error {
	return nil
}

func (a *oAuthService) GetGithubAuthUrl(ctx context.Context, oauthStateString string) string {
	authURL := a.githubOauthConf.AuthCodeURL(oauthStateString)
	fmt.Println(oauthStateString)
	return authURL
}

func (a *oAuthService) GetNewAccessToken(ctx context.Context, refreshToken string) string {
	conf := &oauth2.Config{
		ClientID:     viper.GetString("google.clientID"),
		ClientSecret: viper.GetString("google.clientSecret"),
		Endpoint: oauth2.Endpoint{
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	fmt.Println(refreshToken)

	tokenSource := conf.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})
	newToken, err := tokenSource.Token()
	if err != nil {
		fmt.Println("Error refreshing token:", err)
		return ""
	}

	fmt.Println("New access token: ", newToken.AccessToken)

	return newToken.AccessToken
}

func (a *oAuthService) ExchangeAccessTokenGoogle(ctx context.Context, request model.CallbackGoogleRequest) (*oauth2.Token, error) {
	code := request.AuthCode
	logger.Log.Info(code)

	fmt.Println("code", code)

	if code == "" {
		logger.Log.Warn("Code not found..")
		return nil, errors.New("Code Not Found to provide AccessToken..\n")
	}

	opts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
	}

	token, err := a.googleOauthConf.Exchange(context.Background(), code, opts...)
	if err != nil {
		logger.Log.Error("oauthConfGl.Exchange() failed with " + err.Error() + "\n")
		return nil, err
	}

	fmt.Println("accessToken", token)

	logger.Log.Info("TOKEN>> AccessToken>> " + token.AccessToken)
	logger.Log.Info("TOKEN>> Expiration Time>> " + token.Expiry.String())
	logger.Log.Info("TOKEN>> RefreshToken>> " + token.RefreshToken)

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
	if err != nil {
		logger.Log.Error("Get: " + err.Error() + "\n")
		return nil, err
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Error("ReadAll: " + err.Error() + "\n")
		return nil, err
	}

	logger.Log.Info("parseResponseBody: " + string(response) + "\n")
	fmt.Println(string(response))

	return token, nil
}

func (a *oAuthService) ExchangeAccessTokenGithub(ctx context.Context, request model.CallbackGoogleRequest) (*oauth2.Token, error) {
	code := request.AuthCode
	logger.Log.Info(code)

	fmt.Println("code", code)

	if code == "" {
		logger.Log.Warn("Code not found..")
		return nil, errors.New("Code Not Found to provide AccessToken..\n")
	}

	token, err := a.githubOauthConf.Exchange(context.Background(), code)
	if err != nil {
		logger.Log.Error("oauthConfGl.Exchange() failed with " + err.Error() + "\n")
		return nil, err
	}

	fmt.Println("accessToken", token)

	logger.Log.Info("TOKEN>> AccessToken>> " + token.AccessToken)
	logger.Log.Info("TOKEN>> Expiration Time>> " + token.Expiry.String())
	logger.Log.Info("TOKEN>> RefreshToken>> " + token.RefreshToken)

	return token, nil
}

func (a *oAuthService) RevokeAccessToken(ctx context.Context, accessToken string) error {

	client := &http.Client{}

	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/revoke", strings.NewReader("token="+accessToken))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Token revoked successfully")
	} else {
		fmt.Println("Failed to revoke token. Status:", resp.Status)
	}

	return nil
}
