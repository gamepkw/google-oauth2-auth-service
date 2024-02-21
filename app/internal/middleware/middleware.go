package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
)

type middleware struct {
	SecretKey []byte
}

func InitMiddleware() Middleware {
	return &middleware{}
}

type Middleware interface {
	ExtractJWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc
	GenerateJWTToken(googleClaims string, expiration time.Duration) (string, error)
	CORS(next echo.HandlerFunc) echo.HandlerFunc
}
