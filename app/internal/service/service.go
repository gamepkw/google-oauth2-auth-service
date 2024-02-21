package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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
	GetNewAccessToken(ctx context.Context, refreshToken string) string
	GetFacebookAuthUrl(ctx context.Context, oauthStateString string) string
	GetGithubAuthUrl(ctx context.Context, oauthStateString string) string
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
