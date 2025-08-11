package middleware

import (
	"strconv"
	"vera-identity-service/internal/apperror"
	"vera-identity-service/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware gin.HandlerFunc

func NewAuthMiddleware(config *config.Config) AuthMiddleware {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.Error(apperror.New(apperror.CodeInvalidAuthHeader, "invalid authorization header | header: "+authHeader))
			c.Abort()
			return
		}

		token := authHeader[7:]
		claims := &jwt.RegisteredClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AccessTokenSecret), nil
		})
		if err != nil {
			c.Error(apperror.New(apperror.CodeInvalidAccessToken, "invalid access token | token: "+token))
			c.Abort()
			return
		}

		if claims.Issuer != "identity@vera.sninjo.com" {
			c.Error(apperror.New(apperror.CodeInvalidTokenIssuer, "invalid token issuer | token: "+token))
			c.Abort()
			return
		}

		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			c.Error(apperror.New(apperror.CodeInvalidAccessToken, "invalid access token | token: "+token))
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}
