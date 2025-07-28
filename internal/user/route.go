package user

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/auth/login", loginHandler)
	r.GET("/auth/callback", callbackHandler)
	r.POST("/auth/refresh", refreshHandler)

	rg := r.Group("/users")
	rg.Use(AuthMiddleware())
	{
		rg.GET("", listUsersHandler)
		rg.POST("", createUserHandler)
		rg.PATCH("/:id", updateUserHandler)
		rg.DELETE("/:id", deleteUserHandler)
	}
}
