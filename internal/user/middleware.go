package user

import (
	"strconv"
	"vera-identity-service/internal/apperror"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.Error(apperror.New(apperror.CodeInvalidAuthHeader, "invalid authorization header | header: "+authHeader))
			c.Abort()
			return
		}

		token := authHeader[7:]
		claims, err := parseJWT(token, authConfig.AccessTokenSecret)
		if err != nil {
			c.Error(apperror.New(apperror.CodeInvalidAccessToken, "invalid access token | token: "+token))
			c.Abort()
			return
		}

		userID, _ := strconv.Atoi(claims.Subject)
		user, err := getUserByID(userID)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		} else if user == nil {
			c.Error(apperror.New(apperror.CodeUserNotAuthorized, "user not authorized | id: "+strconv.Itoa(userID)))
			c.Abort()
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}
