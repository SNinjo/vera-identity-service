package user

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	r.GET("/auth/login", loginHandler)
	r.GET("/auth/callback", callbackHandler)
	r.POST("/auth/refresh", refreshHandler)

	rg := r.Group("/")
	rg.Use(AuthMiddleware())
	{
		rg.POST("/auth/verify", verifyHandler)
		rg.GET("/users", listUsersHandler)
		rg.POST("/users", createUserHandler)
		rg.PATCH("/users/:id", updateUserHandler)
		rg.DELETE("/users/:id", deleteUserHandler)
	}
}
