//go:build wireinject
// +build wireinject

package app

import (
	"github.com/sninjo/vera-identity-service/internal/auth"
	"github.com/sninjo/vera-identity-service/internal/config"
	"github.com/sninjo/vera-identity-service/internal/db"
	"github.com/sninjo/vera-identity-service/internal/logger"
	"github.com/sninjo/vera-identity-service/internal/middleware"
	"github.com/sninjo/vera-identity-service/internal/router"
	"github.com/sninjo/vera-identity-service/internal/tool"
	"github.com/sninjo/vera-identity-service/internal/user"

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
