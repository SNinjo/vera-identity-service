package test

import (
	"vera-identity-service/internal/logger"

	"go.uber.org/zap"
)

func SetupLogger() {
	logger.Init(zap.NewNop())
}

func SetupLoggerV() {
	l, _ := zap.NewProduction()
	logger.Init(l)
}
