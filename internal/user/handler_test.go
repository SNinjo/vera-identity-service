package user

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	"vera-identity-service/test"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateUser(email string) error {
	args := m.Called(email)
	return args.Error(0)
}
func (m *MockService) GetUserByID(id int) (*User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}
func (m *MockService) GetUserByEmail(email string) (*User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}
func (m *MockService) GetUsers() ([]User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]User), args.Error(1)
}
func (m *MockService) UpdateUser(id int, email string) error {
	args := m.Called(id, email)
	return args.Error(0)
}
func (m *MockService) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockService) RecordUserLogin(id int, name, picture, loginSub string) error {
	args := m.Called(id, name, picture, loginSub)
	return args.Error(0)
}

func TestHandler_NewHandler_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}

	// Act
	h := NewHandler(mockService)

	// Assert
	assert.IsType(t, &Handler{}, h)
	assert.Equal(t, mockService, h.service)
}

func TestHandler_GetUsers_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	users := []User{
		{
			ID:           1,
			Name:         test.StringPtr("name1"),
			Email:        "user1@example.com",
			Picture:      test.StringPtr("https://example.com/picture1.jpg"),
			LastLoginSub: test.StringPtr("sub1"),
			LastLoginAt:  nil,
			CreatedAt:    time.Unix(1, 0),
			UpdatedAt:    time.Unix(1, 0),
			DeletedAt:    nil,
		},
		{
			ID:           2,
			Name:         test.StringPtr("name2"),
			Email:        "user2@example.com",
			Picture:      test.StringPtr("https://example.com/picture2.jpg"),
			LastLoginSub: test.StringPtr("sub2"),
			LastLoginAt:  test.TimePtr(time.Unix(2, 0)),
			CreatedAt:    time.Unix(2, 0),
			UpdatedAt:    time.Unix(2, 0),
			DeletedAt:    nil,
		},
	}

	mockService.On("GetUsers").Return(users, nil)

	// Act
	handler.GetUsers(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)

	var actual []UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &actual)
	require.NoError(t, err)
	expected := []UserResponse{
		{
			ID:          1,
			Name:        test.StringPtr("name1"),
			Email:       "user1@example.com",
			Picture:     test.StringPtr("https://example.com/picture1.jpg"),
			LastLoginAt: nil,
			CreatedAt:   time.Unix(1, 0).Format(time.RFC3339),
			UpdatedAt:   time.Unix(1, 0).Format(time.RFC3339),
		},
		{
			ID:          2,
			Name:        test.StringPtr("name2"),
			Email:       "user2@example.com",
			Picture:     test.StringPtr("https://example.com/picture2.jpg"),
			LastLoginAt: test.StringPtr(time.Unix(2, 0).Format(time.RFC3339)),
			CreatedAt:   time.Unix(2, 0).Format(time.RFC3339),
			UpdatedAt:   time.Unix(2, 0).Format(time.RFC3339),
		},
	}
	assert.Equal(t, expected, actual)
	mockService.AssertExpectations(t)
}
func TestHandler_GetUsers_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	mockService.On("GetUsers").Return(nil, assert.AnError)

	// Act
	handler.GetUsers(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, assert.AnError, c.Errors[0].Err)
	mockService.AssertExpectations(t)
}

func TestHandler_CreateUser_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	requestBody := RequestBody{
		Email: "user@example.com",
	}
	requestJSON, _ := json.Marshal(requestBody)

	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestJSON))

	mockService.On("CreateUser", requestBody.Email).Return(nil)

	// Act
	handler.CreateUser(c)
	c.Writer.WriteHeaderNow()

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
func TestHandler_CreateUser_InvalidRequestBody(t *testing.T) {
	tests := []struct {
		name          string
		payload       string
		errorContains string
	}{
		{
			name:          "missing email",
			payload:       `{}`,
			errorContains: "Email",
		},
		{
			name:          "invalid email format",
			payload:       `{"email": "not-a-email"}`,
			errorContains: "email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := &MockService{}
			handler := NewHandler(mockService)
			c, w := test.SetupContext()

			c.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(tt.payload)))

			// Act
			handler.CreateUser(c)

			// Assert
			require.Equal(t, http.StatusBadRequest, w.Code)

			var res map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Contains(t, res["error"], "invalid request body")
			assert.Contains(t, res["error"], tt.errorContains)
		})
	}
}
func TestHandler_CreateUser_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	requestBody := RequestBody{
		Email: "user@example.com",
	}
	requestJSON, _ := json.Marshal(requestBody)

	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestJSON))

	mockService.On("CreateUser", requestBody.Email).Return(assert.AnError)

	// Act
	handler.CreateUser(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, assert.AnError, c.Errors[0].Err)
	mockService.AssertExpectations(t)
}

func TestHandler_UpdateUser_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	id := 1
	requestBody := RequestBody{
		Email: "user@example.com",
	}
	requestJSON, _ := json.Marshal(requestBody)

	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestJSON))
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(id)}}

	mockService.On("UpdateUser", id, requestBody.Email).Return(nil)

	// Act
	handler.UpdateUser(c)
	c.Writer.WriteHeaderNow()

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
func TestHandler_UpdateUser_InvalidRequestURI(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}

	// Act
	handler.UpdateUser(c)

	// Assert
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid request uri")
	assert.Contains(t, response["error"], "Int")
}
func TestHandler_UpdateUser_InvalidRequestBody(t *testing.T) {
	id := 1
	tests := []struct {
		name          string
		payload       string
		errorContains string
	}{
		{
			name:          "missing email",
			payload:       `{}`,
			errorContains: "Email",
		},
		{
			name:          "invalid email format",
			payload:       `{"email": "not-a-email"}`,
			errorContains: "email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockService := &MockService{}
			handler := NewHandler(mockService)
			c, w := test.SetupContext()

			c.Request.Body = io.NopCloser(bytes.NewBuffer([]byte(tt.payload)))
			c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(id)}}

			// Act
			handler.UpdateUser(c)

			// Assert
			require.Equal(t, http.StatusBadRequest, w.Code)

			var res map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &res)
			require.NoError(t, err)
			assert.Contains(t, res["error"], "invalid request body")
			assert.Contains(t, res["error"], tt.errorContains)
		})
	}
}
func TestHandler_UpdateUser_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	id := 1
	requestBody := RequestBody{
		Email: "user@example.com",
	}
	requestJSON, _ := json.Marshal(requestBody)

	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestJSON))
	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(id)}}

	mockService.On("UpdateUser", id, requestBody.Email).Return(assert.AnError)

	// Act
	handler.UpdateUser(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, assert.AnError, c.Errors[0].Err)
	mockService.AssertExpectations(t)
}

func TestHandler_DeleteUser_Success(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	id := 1

	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(id)}}

	mockService.On("DeleteUser", id).Return(nil)

	// Act
	handler.DeleteUser(c)
	c.Writer.WriteHeaderNow()

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}
func TestHandler_DeleteUser_InvalidRequestURI(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}

	// Act
	handler.DeleteUser(c)

	// Assert
	require.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid request uri")
	assert.Contains(t, response["error"], "Int")
}
func TestHandler_DeleteUser_ServiceError(t *testing.T) {
	// Arrange
	mockService := &MockService{}
	handler := NewHandler(mockService)
	c, w := test.SetupContext()

	id := 1

	c.Params = gin.Params{{Key: "id", Value: strconv.Itoa(id)}}

	mockService.On("DeleteUser", id).Return(assert.AnError)

	// Act
	handler.DeleteUser(c)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, assert.AnError, c.Errors[0].Err)
	mockService.AssertExpectations(t)
}
