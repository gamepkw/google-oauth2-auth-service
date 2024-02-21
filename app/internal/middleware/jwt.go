package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

type JWTClaims struct {
	GoogleClaims string `json:"googleClaims"`
	jwt.StandardClaims
}

func (m *middleware) GenerateJWTToken(googleClaims string, expiration time.Duration) (string, error) {
	claims := JWTClaims{
		GoogleClaims: googleClaims,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(expiration).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func (m *middleware) ExtractJWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			return c.String(http.StatusUnauthorized, "Invalid token format")
		}

		accessTokenDetail, err := m.getAccessTokenDetail(tokenParts[1])
		if err != nil {
			return c.String(http.StatusUnauthorized, "Cannot validate token")
		}

		if accessTokenDetail.Error.Code == 401 {
			return c.String(http.StatusUnauthorized, "Invalid token")
		} else {
			c.Set("email", accessTokenDetail.Email)
			return next(c)
		}
	}
}

func (m *middleware) getAccessTokenDetail(accessToken string) (AccessTokenDetail, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(accessToken))
	if err != nil {
		return AccessTokenDetail{}, err
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccessTokenDetail{}, err
	}

	var accessTokenDetail AccessTokenDetail
	if err := json.Unmarshal(response, &accessTokenDetail); err != nil {
		return AccessTokenDetail{}, errors.Wrap(err, "failed to unmarshal JSON response")
	}

	return accessTokenDetail, nil
}

type AccessTokenDetail struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Error         Error  `json:"error,omitempty"`
}

type Error struct {
	Code int `json:"code"`
}
