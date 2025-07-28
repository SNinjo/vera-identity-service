package tool

import "github.com/golang-jwt/jwt/v5"

type GenerateJWTRequest struct {
	Claims    jwt.MapClaims `json:"claims"`
	Secret    string        `json:"secret"`
	Issuer    string        `json:"issuer"`
	ExpiredAt int64         `json:"expired_at"`
	IssuedAt  int64         `json:"issued_at"`
}
