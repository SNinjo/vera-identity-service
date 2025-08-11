package router

import (
	"net/http"
	"vera-identity-service/internal/auth"
	"vera-identity-service/internal/middleware"
	"vera-identity-service/internal/tool"
	"vera-identity-service/internal/user"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	httpMiddleware middleware.HTTPMiddleware,
	corsMiddleware middleware.CORSMiddleware,
	authMiddleware middleware.AuthMiddleware,
	toolHandler *tool.Handler,
	authHandler *auth.Handler,
	userHandler *user.Handler,
) *gin.Engine {
	r := gin.New()
	r.Use(
		gin.Recovery(),
		gin.HandlerFunc(httpMiddleware),
		gin.HandlerFunc(corsMiddleware),
	)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})
	r.StaticFile("/docs/swagger.yaml", "./api/swagger.yaml")
	r.StaticFile("/docs", "./api/swagger.html")

	tool.RegisterRoutes(r, toolHandler)
	auth.RegisterRoutes(r, authHandler, authMiddleware)
	user.RegisterRoutes(r, userHandler, authMiddleware)

	return r
}
