package tool

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) getUnixTimestampHandler(c *gin.Context) {
	c.JSON(http.StatusOK, time.Now().Unix())
}

func (h *Handler) generateJWTHandler(c *gin.Context) {
	var req GenerateJWTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	req.Claims["iss"] = req.Issuer
	if req.IssuedAt > 0 {
		req.Claims["iat"] = req.IssuedAt
	}
	if req.ExpiredAt > 0 {
		req.Claims["exp"] = req.ExpiredAt
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, req.Claims).SignedString([]byte(req.Secret))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, token)
}
