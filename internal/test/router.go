package test

import (
	"vera-identity-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(registerRoutes func(r *gin.Engine)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.HTTP())
	registerRoutes(r)
	return r
}
