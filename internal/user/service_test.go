package user

import (
	"testing"
	"time"

	"vera-identity-service/internal/apperror"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}
func (m *MockRepository) GetByID(id int) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}
func (m *MockRepository) GetByEmail(email string) (*User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}
func (m *MockRepository) GetAll() ([]User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]User), args.Error(1)
}
func (m *MockRepository) Update(user *User) error {
	args := m.Called(user)
	return args.Error(0)
}
func (m *MockRepository) SoftDelete(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestService_NewService_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}

	// Act
	s := NewService(mockRepo)

	// Assert
	assert.IsType(t, &service{}, s)
	assert.Equal(t, mockRepo, s.(*service).repo)
}

func TestService_validateEmailUniqueness_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	email := "email@example.com"

	mockRepo.On("GetByEmail", email).Return(nil, nil)

	// Act
	err := service.validateEmailUniqueness(email, nil)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_validateEmailUniqueness_ExcludeID(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	user := &User{ID: 1, Email: "email@example.com"}

	mockRepo.On("GetByEmail", user.Email).Return(user, nil)

	// Act
	err := service.validateEmailUniqueness(user.Email, &user.ID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_validateEmailUniqueness_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	user := &User{ID: 1, Email: "email@example.com"}

	mockRepo.On("GetByEmail", user.Email).Return(user, nil)

	// Act
	err := service.validateEmailUniqueness(user.Email, nil)

	// Assert
	assert.Equal(t, apperror.CodeUserEmailInUse, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_validateEmailUniqueness_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	email := "email@example.com"

	mockRepo.On("GetByEmail", email).Return(nil, assert.AnError)

	// Act
	err := service.validateEmailUniqueness(email, nil)

	// Assert
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_GetUserByID_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	user := &User{ID: 1}

	mockRepo.On("GetByID", user.ID).Return(user, nil)

	// Act
	response, err := service.GetUserByID(user.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, user, response)

	mockRepo.AssertExpectations(t)
}
func TestService_GetUserByID_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	user := &User{ID: 1}

	mockRepo.On("GetByID", user.ID).Return(nil, nil)

	// Act
	response, err := service.GetUserByID(user.ID)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}
func TestService_GetUserByID_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	user := &User{ID: 1}

	mockRepo.On("GetByID", user.ID).Return(nil, assert.AnError)

	// Act
	response, err := service.GetUserByID(user.ID)

	// Assert
	assert.Equal(t, assert.AnError, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}

func TestService_GetUserByEmail_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	user := &User{ID: 1, Email: "email@example.com"}

	mockRepo.On("GetByEmail", user.Email).Return(user, nil)

	// Act
	response, err := service.GetUserByEmail(user.Email)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, user, response)

	mockRepo.AssertExpectations(t)
}
func TestService_GetUserByEmail_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	user := &User{ID: 1, Email: "email@example.com"}

	mockRepo.On("GetByEmail", user.Email).Return(nil, nil)

	// Act
	response, err := service.GetUserByEmail(user.Email)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}
func TestService_GetUserByEmail_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	user := &User{ID: 1, Email: "email@example.com"}

	mockRepo.On("GetByEmail", user.Email).Return(nil, assert.AnError)

	// Act
	response, err := service.GetUserByEmail(user.Email)

	// Assert
	assert.Equal(t, assert.AnError, err)
	assert.Nil(t, response)
	mockRepo.AssertExpectations(t)
}

func TestService_GetUsers_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	users := []User{
		{ID: 1, Email: "email1@example.com"},
		{ID: 2, Email: "email2@example.com"},
	}

	mockRepo.On("GetAll").Return(users, nil)

	// Act
	response, err := service.GetUsers()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, users, response)

	mockRepo.AssertExpectations(t)
}
func TestService_GetUsers_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	mockRepo.On("GetAll").Return(nil, assert.AnError)

	// Act
	response, err := service.GetUsers()

	// Assert
	assert.Equal(t, assert.AnError, err)
	assert.Nil(t, response)

	mockRepo.AssertExpectations(t)
}

func TestService_CreateUser_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	email := "email@example.com"

	mockRepo.On("GetByEmail", email).Return(nil, nil)
	mockRepo.On("Create", &User{Email: email}).Return(nil)

	// Act
	err := service.CreateUser(email)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_CreateUser_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	email := "email@example.com"

	mockRepo.On("GetByEmail", email).Return(&User{Email: email}, nil)

	// Act
	err := service.CreateUser(email)

	// Assert
	assert.Equal(t, apperror.CodeUserEmailInUse, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_CreateUser_CreateError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	email := "email@example.com"

	mockRepo.On("GetByEmail", email).Return(nil, nil)
	mockRepo.On("Create", &User{Email: email}).Return(assert.AnError)

	// Act
	err := service.CreateUser(email)

	// Assert
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_UpdateUser_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1
	oldEmail := "old-email@example.com"
	newEmail := "new-email@example.com"

	mockRepo.On("GetByEmail", newEmail).Return(nil, nil)
	mockRepo.On("GetByID", id).Return(&User{ID: id, Email: oldEmail}, nil)
	mockRepo.On("Update", &User{ID: id, Email: newEmail}).Return(nil)

	// Act
	err := service.UpdateUser(id, newEmail)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_UpdateUser_EmailAlreadyExists(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1
	newEmail := "new-email@example.com"

	mockRepo.On("GetByEmail", newEmail).Return(&User{Email: newEmail}, nil)

	// Act
	err := service.UpdateUser(id, newEmail)

	// Assert
	assert.Equal(t, apperror.CodeUserEmailInUse, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_UpdateUser_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1
	newEmail := "new-email@example.com"

	mockRepo.On("GetByEmail", newEmail).Return(nil, nil)
	mockRepo.On("GetByID", id).Return(nil, nil)

	// Act
	err := service.UpdateUser(id, newEmail)

	// Assert
	assert.Equal(t, apperror.CodeUserNotFound, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_UpdateUser_UpdateError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1
	oldEmail := "old-email@example.com"
	newEmail := "new-email@example.com"

	mockRepo.On("GetByEmail", newEmail).Return(nil, nil)
	mockRepo.On("GetByID", id).Return(&User{ID: id, Email: oldEmail}, nil)
	mockRepo.On("Update", &User{ID: id, Email: newEmail}).Return(assert.AnError)

	// Act
	err := service.UpdateUser(id, newEmail)

	// Assert
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_DeleteUser_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1

	mockRepo.On("GetByID", id).Return(&User{ID: id}, nil)
	mockRepo.On("SoftDelete", id).Return(nil)

	// Act
	err := service.DeleteUser(id)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
func TestService_DeleteUser_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1

	mockRepo.On("GetByID", id).Return(nil, nil)

	// Act
	err := service.DeleteUser(id)

	// Assert
	assert.Equal(t, apperror.CodeUserNotFound, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_DeleteUser_SoftDeleteError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1

	mockRepo.On("GetByID", id).Return(&User{ID: id}, nil)
	mockRepo.On("SoftDelete", id).Return(assert.AnError)

	// Act
	err := service.DeleteUser(id)

	// Assert
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

func TestService_RecordUserLogin_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1
	name := "John Doe"
	picture := "https://example.com/picture.jpg"
	loginSub := "google"

	mockRepo.On("GetByID", id).Return(&User{ID: id}, nil)
	mockRepo.On("Update", mock.MatchedBy(func(u *User) bool {
		return u != nil &&
			u.LastLoginAt != nil &&
			u.ID == id &&
			*u.Name == name &&
			*u.Picture == picture &&
			*u.LastLoginSub == loginSub &&
			time.Since(*u.LastLoginAt) < time.Second
	})).Return(nil)

	// Act
	err := service.RecordUserLogin(id, name, picture, loginSub)

	// Assert
	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_RecordUserLogin_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1
	name := "John Doe"
	picture := "https://example.com/picture.jpg"
	loginSub := "google"

	mockRepo.On("GetByID", id).Return(nil, nil)

	// Act
	err := service.RecordUserLogin(id, name, picture, loginSub)

	// Assert
	assert.Equal(t, apperror.CodeUserNotFound, err.(*apperror.AppError).Code)
	mockRepo.AssertExpectations(t)
}
func TestService_RecordUserLogin_UpdateError(t *testing.T) {
	// Arrange
	mockRepo := &MockRepository{}
	service := &service{repo: mockRepo}

	id := 1
	name := "John Doe"
	picture := "https://example.com/picture.jpg"
	loginSub := "google"

	mockRepo.On("GetByID", id).Return(&User{ID: id}, nil)
	mockRepo.On("Update", mock.MatchedBy(func(u *User) bool {
		return u != nil &&
			u.LastLoginAt != nil &&
			u.ID == id &&
			*u.Name == name &&
			*u.Picture == picture &&
			*u.LastLoginSub == loginSub &&
			time.Since(*u.LastLoginAt) < time.Second
	})).Return(assert.AnError)

	// Act
	err := service.RecordUserLogin(id, name, picture, loginSub)

	// Assert
	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}
