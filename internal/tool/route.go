package tool

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", checkHealthHandler)
	r.GET("/unix-timestamp", getUnixTimestampHandler)
	r.POST("/jwt", generateJWTHandler)
}
