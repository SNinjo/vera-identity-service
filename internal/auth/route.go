package auth

import (
	"vera-identity-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, handler *Handler, authMiddleware middleware.AuthMiddleware) {
	r.GET("/auth/login", handler.Login)
	r.GET("/auth/callback", handler.Callback)
	r.POST("/auth/refresh", handler.Refresh)
	r.POST("/auth/verify", gin.HandlerFunc(authMiddleware), handler.Verify)
}
