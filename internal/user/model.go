package user

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           int            `json:"id" gorm:"primaryKey;autoIncrement"`
	Email        string         `json:"email" gorm:"type:varchar(255);not null"`
	Name         *string        `json:"name" gorm:"type:varchar(255)"`
	Picture      string         `json:"picture" gorm:"type:varchar(255)"`
	LastLoginSub *string        `json:"last_login_sub,omitempty" gorm:"type:varchar(255)"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty" gorm:"type:timestamp with time zone"`
	CreatedAt    time.Time      `json:"created_at" gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"type:timestamp with time zone;default:CURRENT_TIMESTAMP"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"type:timestamp with time zone;index"`
}

func (User) TableName() string {
	return "users"
}
