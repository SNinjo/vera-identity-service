package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"vera-identity-service/internal/apperror"
	"vera-identity-service/internal/user"
	"vera-identity-service/test"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GetOAuthLoginURL() string {
	args := m.Called()
	return args.String(0)
}
func (m *MockAuthService) GetOAuthIDTokenClaims(code string) (*OAuthIDTokenClaims, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OAuthIDTokenClaims), args.Error(1)
}
func (m *MockAuthService) NewAccessToken(id int, name, email, picture string) (string, error) {
	args := m.Called(id, name, email, picture)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}
func (m *MockAuthService) NewRefreshToken(id int) (string, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}
func (m *MockAuthService) ParseAccessToken(token string) (*TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenClaims), args.Error(1)
}
func (m *MockAuthService) ParseRefreshToken(token string) (*TokenClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*TokenClaims), args.Error(1)
}

func TestHandler_NewHandler_Success(t *testing.T) {
	// Arrange
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	config := NewMockConfig("")

	// Act
	h := NewHandler(config, mockAuthService, mockUserService)

	// Assert
	assert.IsType(t, &Handler{}, h)
	assert.Equal(t, config, h.config)
	assert.Equal(t, mockAuthService, h.authService)
	assert.Equal(t, mockUserService, h.userService)
}

func TestHandler_Login_Success(t *testing.T) {
	// Arrange
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	handler := NewHandler(config, mockAuthService, mockUserService)
	c, w := test.SetupContext()

	loginURL := "http://mock-oauth-url/auth"

	mockAuthService.On("GetOAuthLoginURL").Return(loginURL)

	// Act
	handler.Login(c)

	// Assert
	mockAuthService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)

	require.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, loginURL, w.Header().Get("Location"))
}

func TestHandler_Callback_Success(t *testing.T) {
	// Arrange
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	handler := NewHandler(config, mockAuthService, mockUserService)
	c, w := test.SetupContext()

	code := "mock-code"
	idTokenClaims := &OAuthIDTokenClaims{
		Name:    "mock-name",
		Email:   "mock-email",
		Picture: "mock-picture",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "mock-subject",
		},
	}
	user := &user.User{
		ID:      1,
		Name:    test.StringPtr("mock-name2"),
		Email:   "mock-email2",
		Picture: test.StringPtr("mock-picture2"),
	}
	mockAccessToken := "mock-access-token"
	mockRefreshToken := "mock-refresh-token"

	c.Request.URL.RawQuery = fmt.Sprintf("code=%s", code)

	mockAuthService.On("GetOAuthIDTokenClaims", code).Return(idTokenClaims, nil)
	mockUserService.On("GetUserByEmail", idTokenClaims.Email).Return(user, nil)
	mockUserService.On("RecordUserLogin", user.ID, idTokenClaims.Name, idTokenClaims.Picture, idTokenClaims.Subject).Return(nil)
	mockAuthService.On("NewAccessToken", user.ID, idTokenClaims.Name, idTokenClaims.Email, idTokenClaims.Picture).Return(mockAccessToken, nil)
	mockAuthService.On("NewRefreshToken", user.ID).Return(mockRefreshToken, nil)

	// Act
	handler.Callback(c)

	// Assert
	mockAuthService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	require.Equal(t, http.StatusFound, w.Code)

	cookies := w.Result().Cookies()
	var refreshTokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshTokenCookie = cookie
			break
		}
	}
	assert.Equal(t, mockRefreshToken, refreshTokenCookie.Value)
	assert.Equal(t, "/", refreshTokenCookie.Path)
	assert.Empty(t, refreshTokenCookie.Domain)
	assert.Equal(t, int(config.RefreshTokenTTL.Seconds()), refreshTokenCookie.MaxAge)
	assert.True(t, refreshTokenCookie.HttpOnly)
	assert.True(t, refreshTokenCookie.Secure)

	location := w.Header().Get("Location")
	actual, err := url.Parse(location)
	require.NoError(t, err)
	expected, err := url.Parse(config.SiteURL + "?access_token=" + mockAccessToken)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
func TestHandler_Callback_UserNotFound(t *testing.T) {
	// Arrange
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	handler := NewHandler(config, mockAuthService, mockUserService)
	c, w := test.SetupContext()

	code := "mock-code"
	idTokenClaims := &OAuthIDTokenClaims{
		Name:    "mock-name",
		Email:   "mock-email",
		Picture: "mock-picture",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "mock-subject",
		},
	}

	c.Request.URL.RawQuery = fmt.Sprintf("code=%s", code)

	mockAuthService.On("GetOAuthIDTokenClaims", code).Return(idTokenClaims, nil)
	mockUserService.On("GetUserByEmail", idTokenClaims.Email).Return(nil, nil)

	// Act
	handler.Callback(c)

	// Assert
	mockAuthService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, apperror.CodeUserNotAuthorized, c.Errors[0].Err.(*apperror.AppError).Code)
}

func TestHandler_Refresh_Success(t *testing.T) {
	// Arrange
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	handler := NewHandler(config, mockAuthService, mockUserService)
	c, w := test.SetupContext()

	mockAccessToken := "mock-access-token"
	mockRefreshToken := "mock-refresh-token"
	user := &user.User{
		ID:      1,
		Name:    test.StringPtr("mock-name"),
		Email:   "mock-email",
		Picture: test.StringPtr("mock-picture"),
	}
	claims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(user.ID),
		},
	}

	c.Request.AddCookie(&http.Cookie{Name: "refresh_token", Value: mockRefreshToken})

	mockAuthService.On("ParseRefreshToken", mockRefreshToken).Return(claims, nil)
	mockUserService.On("GetUserByID", user.ID).Return(user, nil)
	mockAuthService.On("NewAccessToken", user.ID, *user.Name, user.Email, *user.Picture).Return(mockAccessToken, nil)

	// Act
	handler.Refresh(c)

	// Assert
	mockAuthService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	require.Equal(t, http.StatusOK, w.Code)

	var resp TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, mockAccessToken, resp.AccessToken)
}
func TestHandler_Refresh_InvalidRefreshToken(t *testing.T) {
	// Arrange
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	handler := NewHandler(config, mockAuthService, mockUserService)
	c, w := test.SetupContext()

	mockRefreshToken := "mock-refresh-token"

	c.Request.AddCookie(&http.Cookie{Name: "refresh_token", Value: mockRefreshToken})

	mockAuthService.On("ParseRefreshToken", mockRefreshToken).Return(nil, assert.AnError)

	// Act
	handler.Refresh(c)

	// Assert
	mockAuthService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, apperror.CodeInvalidRefreshToken, c.Errors[0].Err.(*apperror.AppError).Code)
}
func TestHandler_Refresh_UserNotFound(t *testing.T) {
	// Arrange
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	handler := NewHandler(config, mockAuthService, mockUserService)
	c, w := test.SetupContext()

	mockRefreshToken := "mock-refresh-token"
	user := &user.User{ID: 1}
	claims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(user.ID),
		},
	}

	c.Request.AddCookie(&http.Cookie{Name: "refresh_token", Value: mockRefreshToken})

	mockAuthService.On("ParseRefreshToken", mockRefreshToken).Return(claims, nil)
	mockUserService.On("GetUserByID", user.ID).Return(nil, nil)

	// Act
	handler.Refresh(c)

	// Assert
	mockAuthService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, c.Errors, 1)
	assert.Equal(t, apperror.CodeUserNotAuthorized, c.Errors[0].Err.(*apperror.AppError).Code)
}

func TestHandler_Verify_Success(t *testing.T) {
	// Arrange
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	handler := NewHandler(config, mockAuthService, mockUserService)
	c, w := test.SetupContext()

	// Act
	handler.Verify(c)
	c.Writer.WriteHeaderNow()

	// Assert
	mockAuthService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	require.Equal(t, http.StatusNoContent, w.Code)
}
