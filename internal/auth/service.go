package auth

import (
	"context"
	"errors"
	"strconv"
	"time"

	"vera-identity-service/internal/apperror"
	"vera-identity-service/internal/config"
	"vera-identity-service/internal/user"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type Service interface {
	GetOAuthLoginURL() string
	GetOAuthIDTokenClaims(code string) (*OAuthIDTokenClaims, error)
	NewAccessToken(id int, name, email, picture string) (string, error)
	NewRefreshToken(id int) (string, error)
	ParseAccessToken(token string) (*TokenClaims, error)
	ParseRefreshToken(token string) (*TokenClaims, error)
}

type service struct {
	config      *config.Config
	userService user.Service
}

func NewService(config *config.Config, userService user.Service) Service {
	return &service{config: config, userService: userService}
}

func (s *service) GetOAuthLoginURL() string {
	// TODO: generate and store CSRF state securely
	return s.config.OAuth2.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

type IDTokenClaims struct {
	Sub     string
	Email   string
	Picture string
	Name    string
}

func (s *service) GetOAuthIDTokenClaims(code string) (*OAuthIDTokenClaims, error) {
	oauthToken, err := s.config.OAuth2.Exchange(context.Background(), code)
	if err != nil {
		return nil, apperror.New(apperror.CodeInvalidOAuthCode, "failed to exchange OAuth code | code: "+code)
	}

	idToken, _ := oauthToken.Extra("id_token").(string)
	claims := &OAuthIDTokenClaims{}
	_, _, err = new(jwt.Parser).ParseUnverified(idToken, claims)
	if err != nil {
		return nil, apperror.New(apperror.CodeInvalidOAuthIdToken, "failed to parse id_token | id_token: "+idToken)
	}

	if claims.Subject == "" || claims.Name == "" || claims.Email == "" || claims.Picture == "" {
		return nil, apperror.New(
			apperror.CodeMissingUserInfo,
			"missing sub, name, email, or picture "+
				"| subject: "+claims.Subject+" | name: "+claims.Name+" | email: "+claims.Email+" | picture: "+claims.Picture,
		)
	}

	return claims, nil
}

func (s *service) NewAccessToken(id int, name, email, picture string) (string, error) {
	claims := TokenClaims{
		Name:    name,
		Email:   email,
		Picture: picture,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(id),
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.AccessTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.config.AccessTokenSecret)
}

func (s *service) NewRefreshToken(id int) (string, error) {
	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(id),
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.RefreshTokenTTL)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.config.RefreshTokenSecret)
}

func (s *service) ParseAccessToken(token string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return s.config.AccessTokenSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims.Issuer != "identity@vera.sninjo.com" {
		return nil, errors.New("invalid token issuer")
	}

	return claims, nil
}

func (s *service) ParseRefreshToken(token string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return s.config.RefreshTokenSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims.Issuer != "identity@vera.sninjo.com" {
		return nil, errors.New("invalid token issuer")
	}

	return claims, nil
}
