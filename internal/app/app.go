package app

import (
	"vera-identity-service/internal/auth"
	"vera-identity-service/internal/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Config      *config.Config
	Router      *gin.Engine
	DB          *gorm.DB
	Logger      *zap.Logger
	AuthService auth.Service
}

func NewApp(
	config *config.Config,
	router *gin.Engine,
	db *gorm.DB,
	logger *zap.Logger,
	authService auth.Service,
) *App {
	return &App{
		Config:      config,
		Router:      router,
		DB:          db,
		Logger:      logger,
		AuthService: authService,
	}
}

func (a *App) Close() {
	a.Logger.Sync()
}

func (a *App) Run() {
	if err := a.Router.Run(a.Config.Domain + ":" + a.Config.Port); err != nil {
		a.Logger.Fatal("failed to run server", zap.Error(err))
	}
}
