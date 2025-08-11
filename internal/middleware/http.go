package middleware

import (
	"bytes"
	"io"
	"time"
	"vera-identity-service/internal/apperror"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type HTTPMiddleware gin.HandlerFunc

func logRequest(c *gin.Context, logger *zap.Logger, requestID string) {
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
	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", method),
		zap.String("path", path),
		zap.String("query", query),
		zap.String("body", string(bodyBytes)),
		zap.Strings("cookies", cookieStrs),
		zap.String("auth_header", authHeader),
	}
	logger.Info("request started", fields...)
}

func NewHTTPMiddleware(logger *zap.Logger) HTTPMiddleware {
	return func(c *gin.Context) {
		start := time.Now()
		requestID := uuid.New().String()
		logRequest(c, logger, requestID)

		c.Next()

		err := c.Errors.Last()
		var appErr *apperror.AppError
		if err != nil {
			appErr = apperror.FromError(err)
			c.JSON(appErr.Status, appErr.Response)
		}

		f := []zap.Field{
			zap.String("request_id", requestID),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("duration_ms", time.Since(start).Milliseconds()),
			zap.Error(appErr),
		}
		logger.Info("request completed", f...)
	}
}
