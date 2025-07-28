package logger

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func Init(logger *zap.Logger) {
	Logger = logger
}
