package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	BaseURL     string
	Domain      string
	Port        string
	DatabaseURL string
	SiteURL     string

	GoogleClientID     string
	GoogleClientSecret string
	OAuth2             *oauth2.Config

	AccessTokenTTL     time.Duration
	AccessTokenSecret  []byte
	RefreshTokenTTL    time.Duration
	RefreshTokenSecret []byte
}

func NewConfig(logger *zap.Logger) *Config {
	err := godotenv.Load()
	if err != nil {
		logger.Warn("No .env file found or error loading .env file", zap.Error(err))
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
		logger.Fatal("Invalid ACCESS_TOKEN_TTL", zap.Error(err))
	}
	refreshTokenTTL, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_TTL"))
	if err != nil {
		logger.Fatal("Invalid REFRESH_TOKEN_TTL", zap.Error(err))
	}

	baseURL := os.Getenv("BASE_URL")
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	oauth2Config := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  baseURL + "/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return &Config{
		BaseURL:     os.Getenv("BASE_URL"),
		Domain:      domain,
		Port:        port,
		DatabaseURL: os.Getenv("DATABASE_URL"),
		SiteURL:     os.Getenv("SITE_URL"),

		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		OAuth2:             oauth2Config,

		AccessTokenTTL:     accessTokenTTL,
		AccessTokenSecret:  []byte(os.Getenv("ACCESS_TOKEN_SECRET")),
		RefreshTokenTTL:    refreshTokenTTL,
		RefreshTokenSecret: []byte(os.Getenv("REFRESH_TOKEN_SECRET")),
	}
}
