//go:build wireinject
// +build wireinject

package app

import (
	"vera-identity-service/internal/auth"
	"vera-identity-service/internal/config"
	"vera-identity-service/internal/db"
	"vera-identity-service/internal/logger"
	"vera-identity-service/internal/middleware"
	"vera-identity-service/internal/router"
	"vera-identity-service/internal/tool"
	"vera-identity-service/internal/user"

	"github.com/google/wire"
)

func InitApp() (*App, error) {
	wire.Build(
		logger.NewLogger,
		config.NewConfig,
		db.NewDatabase,
		middleware.NewHTTPMiddleware,
		middleware.NewCORSMiddleware,
		middleware.NewAuthMiddleware,
		tool.NewHandler,
		auth.NewService,
		auth.NewHandler,
		user.NewRepository,
		user.NewService,
		user.NewHandler,
		router.NewRouter,
		NewApp,
	)
	return &App{}, nil
}
