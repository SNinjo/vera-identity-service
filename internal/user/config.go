package user

import (
	"time"

	"golang.org/x/oauth2"
)

type AuthConfig struct {
	BaseURL            string
	OAuthClientID      string
	OAuthClientSecret  string
	OAuthEndpoint      oauth2.Endpoint
	AccessTokenSecret  string
	AccessTokenTTL     time.Duration
	RefreshTokenSecret string
	RefreshTokenTTL    time.Duration
}

var oauthConfig *oauth2.Config
var authConfig *AuthConfig

func Init(config *AuthConfig) {
	authConfig = config
	oauthConfig = &oauth2.Config{
		ClientID:     config.OAuthClientID,
		ClientSecret: config.OAuthClientSecret,
		RedirectURL:  config.BaseURL + "/auth/callback",
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     config.OAuthEndpoint,
	}
}
