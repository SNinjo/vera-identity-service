package user

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type requestUri struct {
	ID int `uri:"id" binding:"required,min=1"`
}

type requestBody struct {
	Email string `json:"email" binding:"required,email,max=255"`
}

type tokenClaims struct {
	Email   string `json:"email,omitempty"`
	Picture string `json:"picture,omitempty"`
	jwt.RegisteredClaims
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

type userResponse struct {
	ID          int        `json:"id"`
	Email       string     `json:"email"`
	Picture     string     `json:"picture"`
	LastLoginAt *time.Time `json:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func newUserResponse(u *User) *userResponse {
	return &userResponse{
		ID:          u.ID,
		Email:       u.Email,
		Picture:     u.Picture,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}
