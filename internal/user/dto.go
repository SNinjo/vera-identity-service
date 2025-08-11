package user

import "time"

type RequestUri struct {
	ID int `uri:"id" binding:"required,min=1"`
}

type RequestBody struct {
	Email string `json:"email" binding:"email,max=255"`
}

type UserResponse struct {
	ID          int     `json:"id"`
	Name        *string `json:"name"`
	Email       string  `json:"email"`
	Picture     *string `json:"picture"`
	LastLoginAt *string `json:"last_login_at"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func newUserResponse(u *User) *UserResponse {
	var lastLoginAt *string
	if u.LastLoginAt != nil {
		t := u.LastLoginAt.Format(time.RFC3339)
		lastLoginAt = &t
	}
	return &UserResponse{
		ID:          u.ID,
		Name:        u.Name,
		Email:       u.Email,
		Picture:     u.Picture,
		LastLoginAt: lastLoginAt,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   u.UpdatedAt.Format(time.RFC3339),
	}
}
