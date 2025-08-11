package user

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           int        `gorm:"primaryKey;autoIncrement"`
	Name         *string    `gorm:"type:varchar(255)"`
	Email        string     `gorm:"type:varchar(255);not null"`
	Picture      *string    `gorm:"type:varchar(255)"`
	LastLoginSub *string    `gorm:"type:varchar(255)"`
	LastLoginAt  *time.Time `gorm:"type:timestamptz"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;not null"`
	UpdatedAt    time.Time  `gorm:"type:timestamptz;not null"`
	DeletedAt    *time.Time `gorm:"type:timestamptz;index"`
}

func (User) TableName() string {
	return "users"
}

type Repository interface {
	GetByID(id int) (*User, error)
	GetByEmail(email string) (*User, error)
	GetAll() ([]User, error)
	Create(user *User) error
	Update(user *User) error
	SoftDelete(id int) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetByID(id int) (*User, error) {
	var user User
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetAll() ([]User, error) {
	var users []User
	err := r.db.Where("deleted_at IS NULL").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *repository) Create(user *User) error {
	user.ID = 0
	user.CreatedAt = time.Now().Local()
	user.UpdatedAt = time.Now().Local()
	return r.db.Create(user).Error
}

func (r *repository) Update(user *User) error {
	existingUser, err := r.GetByID(user.ID)
	if err != nil {
		return err
	}
	if existingUser == nil {
		return errors.New("user not exist")
	}

	user.UpdatedAt = time.Now().Local()
	return r.db.Save(user).Error
}

func (r *repository) SoftDelete(id int) error {
	return r.db.Model(&User{}).Where("id = ?", id).Update("deleted_at", time.Now().Local()).Error
}
