package middleware

import (
	"bytes"
	"io"
	"time"
	"vera-identity-service/internal/apperror"
	"vera-identity-service/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func logRequest(c *gin.Context, f *[]zap.Field) {
	method := c.Request.Method
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	cookies := c.Request.Cookies()
	cookieStrs := make([]string, len(cookies))
	for i, c := range c.Request.Cookies() {
		cookieStrs[i] = c.Name + "=" + c.Value
	}
	authHeader := c.GetHeader("Authorization")
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}
	*f = append(*f,
		zap.String("method", method),
		zap.String("path", path),
		zap.String("query", query),
		zap.String("request_body", string(bodyBytes)),
		zap.Strings("cookies", cookieStrs),
		zap.String("auth_header", authHeader),
	)
}

func HTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		err := c.Errors.Last()
		var appErr *apperror.AppError
		if err != nil {
			appErr = apperror.FromError(err)
			c.JSON(appErr.Status, appErr.Response)
		}

		fields := []zap.Field{
			zap.Int("status", c.Writer.Status()),
			zap.Int64("duration_ms", duration.Milliseconds()),
			zap.Error(appErr),
		}
		logRequest(c, &fields)
		logger.Logger.Info("request completed", fields...)
	}
}
