package user

import (
	"strconv"
	"time"

	"vera-identity-service/internal/apperror"
)

type Service interface {
	GetUserByID(id int) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUsers() ([]User, error)
	CreateUser(email string) error
	UpdateUser(id int, email string) error
	DeleteUser(id int) error
	RecordUserLogin(id int, name, picture, loginSub string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) validateEmailUniqueness(email string, excludeID *int) error {
	existingUser, err := s.repo.GetByEmail(email)
	if err != nil {
		return err
	}
	if existingUser != nil && (excludeID == nil || existingUser.ID != *excludeID) {
		return apperror.New(apperror.CodeUserEmailInUse, "user email already in use | email: "+email)
	}
	return nil
}

func (s *service) GetUserByID(id int) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *service) GetUserByEmail(email string) (*User, error) {
	return s.repo.GetByEmail(email)
}

func (s *service) GetUsers() ([]User, error) {
	return s.repo.GetAll()
}

func (s *service) CreateUser(email string) error {
	if err := s.validateEmailUniqueness(email, nil); err != nil {
		return err
	}

	user := &User{Email: email}
	if err := s.repo.Create(user); err != nil {
		return err
	}
	return nil
}

func (s *service) UpdateUser(id int, email string) error {
	if err := s.validateEmailUniqueness(email, &id); err != nil {
		return err
	}

	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.New(apperror.CodeUserNotFound, "user not found | id: "+strconv.Itoa(id))
	}

	if email != "" {
		user.Email = email
	}
	if err := s.repo.Update(user); err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteUser(id int) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.New(apperror.CodeUserNotFound, "user not found | id: "+strconv.Itoa(id))
	}

	return s.repo.SoftDelete(id)
}

func (s *service) RecordUserLogin(id int, name, picture, loginSub string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.New(apperror.CodeUserNotFound, "user not found | id: "+strconv.Itoa(id))
	}

	user.Name = &name
	user.Picture = &picture
	user.LastLoginSub = &loginSub
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.repo.Update(user); err != nil {
		return err
	}
	return nil
}
