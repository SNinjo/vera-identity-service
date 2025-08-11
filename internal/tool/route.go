package tool

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *Handler) {
	r.GET("/unix-timestamp", h.getUnixTimestampHandler)
	r.POST("/jwt", h.generateJWTHandler)
}
