package user

import (
	"github.com/sninjo/vera-identity-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, handler *Handler, authMiddleware middleware.AuthMiddleware) {
	g := r.Group("/users")
	g.Use(gin.HandlerFunc(authMiddleware))
	{
		g.GET("", handler.GetUsers)
		g.POST("", handler.CreateUser)
		g.PATCH("/:id", handler.UpdateUser)
		g.DELETE("/:id", handler.DeleteUser)
	}
}
