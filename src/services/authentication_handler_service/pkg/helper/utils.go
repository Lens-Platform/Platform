package helper

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JwtCustomClaims struct {
	Id string `json:"id"`
	jwt.StandardClaims
}

type TokenValidationResponse struct {
	Id        string    `json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}
