package user

import (
	"context"
	"strconv"
	"time"

	"vera-identity-service/internal/apperror"
	"vera-identity-service/internal/db"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

func getOAuthLoginURL() string {
	// TODO: generate and store CSRF state securely
	return oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

type oauthClaims struct {
	Email   string `json:"email"`
	Sub     string `json:"sub"`
	Picture string `json:"picture"`
	Name    string `json:"name"`
	jwt.RegisteredClaims
}
type oauthResponse struct {
	Sub     string
	Email   string
	Picture string
	Name    string
}

func handleOAuthCallback(code string) (*oauthResponse, error) {
	oauthToken, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, apperror.New(apperror.CodeInvalidOAuthCode, "failed to exchange OAuth code | code: "+code)
	}
	idToken, _ := oauthToken.Extra("id_token").(string)
	claims := &oauthClaims{}
	_, _, err = new(jwt.Parser).ParseUnverified(idToken, claims)
	if err != nil {
		return nil, apperror.New(apperror.CodeInvalidOAuthIdToken, "failed to parse id_token | id_token: "+idToken)
	}
	if claims.Sub == "" || claims.Email == "" || claims.Picture == "" {
		return nil, apperror.New(
			apperror.CodeMissingUserInfo,
			"missing sub, email, or picture | sub: "+claims.Sub+" | email: "+claims.Email+" | picture: "+claims.Picture,
		)
	}
	return &oauthResponse{
		Sub:     claims.Sub,
		Email:   claims.Email,
		Picture: claims.Picture,
		Name:    claims.Name,
	}, nil
}

func newJWT(user *User, secret string, ttl time.Duration) (string, error) {
	claims := tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(user.ID),
			Issuer:    "identity@vera.sninjo.com",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	if user.Email != "" {
		claims.Email = user.Email
	}
	if user.Name != nil {
		claims.Name = *user.Name
	}
	if user.Picture != "" {
		claims.Picture = user.Picture
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func parseJWT(token string, secret string) (*tokenClaims, error) {
	claims := &tokenClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims.Issuer != "identity@vera.sninjo.com" {
		return nil, apperror.New(apperror.CodeInvalidTokenIssuer, "invalid token issuer | token: "+token)
	}

	return claims, nil
}

func getUserByID(id int) (*User, error) {
	var u User
	err := db.DB.Where("id = ?", id).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func getUserByEmail(email string) (*User, error) {
	var u User
	err := db.DB.Where("email = ?", email).First(&u).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func getUsers() ([]User, error) {
	var users []User
	err := db.DB.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func createUser(user *User) error {
	return db.DB.Create(user).Error
}

func updateUser(id int, updates *User) (*User, error) {
	existingUser, err := getUserByID(id)
	if err != nil {
		return nil, err
	} else if existingUser == nil {
		return nil, apperror.New(apperror.CodeUserNotFound, "user not found | id: "+strconv.Itoa(id))
	}

	err = db.DB.Model(&User{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		return nil, err
	}

	return getUserByID(id)
}

func deleteUser(id int) error {
	return db.DB.Delete(&User{ID: id}).Error
}
