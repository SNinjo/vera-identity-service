package main

import (
	"log"

	"go.uber.org/zap"
	"golang.org/x/oauth2/google"

	"vera-identity-service/internal/config"
	"vera-identity-service/internal/db"
	customLogger "vera-identity-service/internal/logger"
	"vera-identity-service/internal/middleware"
	"vera-identity-service/internal/tool"
	"vera-identity-service/internal/user"

	"github.com/gin-gonic/gin"
)

func main() {
	logger, err := zap.NewProduction()
	customLogger.Init(logger)
	if err != nil {
		log.Fatal("failed to initialize logger", zap.Error(err))
	}
	defer logger.Sync()
	cfg := config.Load()
	user.Init(&user.AuthConfig{
		BaseURL:            cfg.BaseURL,
		FrontendURL:        cfg.FrontendURL,
		OAuthClientID:      cfg.GoogleClientID,
		OAuthClientSecret:  cfg.GoogleClientSecret,
		OAuthEndpoint:      google.Endpoint,
		AccessTokenSecret:  cfg.AccessTokenSecret,
		AccessTokenTTL:     cfg.AccessTokenTTL,
		RefreshTokenSecret: cfg.RefreshTokenSecret,
		RefreshTokenTTL:    cfg.RefreshTokenTTL,
	})
	_, err = db.Init(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("failed to initialize database", zap.Error(err))
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.HTTP())
	r.Use(middleware.CORS([]string{cfg.FrontendURL}))
	r.StaticFile("/docs/swagger.yaml", "./api/swagger.yaml")
	r.StaticFile("/docs", "./api/swagger.html")
	tool.RegisterRoutes(r)
	user.RegisterRoutes(r)

	if err := r.Run(cfg.Domain + ":" + cfg.Port); err != nil {
		logger.Fatal("failed to run server", zap.Error(err))
	}
}
