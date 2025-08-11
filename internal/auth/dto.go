package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type OAuthIDTokenClaims struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
	jwt.RegisteredClaims
}

type TokenClaims struct {
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	Picture string `json:"picture,omitempty"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}
