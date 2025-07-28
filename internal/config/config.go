package config

import (
	"os"
	"time"
	"vera-identity-service/internal/logger"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	BaseURL     string
	Domain      string
	Port        string
	DatabaseURL string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	AccessTokenTTL     time.Duration
	AccessTokenSecret  string
	RefreshTokenTTL    time.Duration
	RefreshTokenSecret string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Logger.Warn("No .env file found or error loading .env file", zap.Error(err))
	}
	var domain string
	mode := os.Getenv("GIN_MODE")
	if mode == "release" {
		domain = "0.0.0.0"
	} else {
		domain = "localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	accessTokenTTL, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_TTL"))
	if err != nil {
		logger.Logger.Fatal("Invalid ACCESS_TOKEN_TTL", zap.Error(err))
	}
	refreshTokenTTL, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_TTL"))
	if err != nil {
		logger.Logger.Fatal("Invalid REFRESH_TOKEN_TTL", zap.Error(err))
	}
	return &Config{
		BaseURL:     os.Getenv("BASE_URL"),
		Domain:      domain,
		Port:        port,
		DatabaseURL: os.Getenv("DB_URL"),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),

		AccessTokenTTL:     accessTokenTTL,
		AccessTokenSecret:  os.Getenv("ACCESS_TOKEN_SECRET"),
		RefreshTokenTTL:    refreshTokenTTL,
		RefreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET"),
	}
}
