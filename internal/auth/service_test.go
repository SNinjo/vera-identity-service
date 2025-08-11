package auth

import (
	"net/url"
	"testing"
	"time"
	"vera-identity-service/test"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_NewService_Success(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")

	// Act
	s := NewService(config, mockUserService)

	// Assert
	assert.IsType(t, &service{}, s)
	assert.Equal(t, mockUserService, s.(*service).userService)
	assert.Equal(t, config, s.(*service).config)
}

func TestService_GetOAuthLoginURL_Success(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("http://mock-oauth-url")
	service := NewService(config, mockUserService)

	// Act
	actualURLStr := service.GetOAuthLoginURL()

	// Assert
	mockUserService.AssertExpectations(t)

	actualURL, err := url.Parse(actualURLStr)
	require.NoError(t, err)
	expectedURL, err := url.Parse(
		"http://mock-oauth-url/auth" +
			"?access_type=offline" +
			"&client_id=mock-client-id" +
			"&redirect_uri=http%3A%2F%2Fmock-base-url%2Fauth%2Fcallback" +
			"&response_type=code" +
			"&scope=openid+email+profile" +
			"&state=state",
	)
	require.NoError(t, err)
	assert.Equal(t, expectedURL, actualURL)
}

func TestService_GetOAuthIDTokenClaims_Success(t *testing.T) {
	// Arrange
	oauthAPI := test.SetupOAuthAPI()
	mockUserService := &MockUserService{}
	config := NewMockConfig(oauthAPI.URL)
	service := NewService(config, mockUserService)

	// Act
	claims, err := service.GetOAuthIDTokenClaims(oauthAPI.AuthorizationCode)

	// Assert
	mockUserService.AssertExpectations(t)
	require.NoError(t, err)

	assert.Equal(t, oauthAPI.IDTokenClaims.Subject, claims.Subject)
	assert.Equal(t, oauthAPI.IDTokenClaims.Name, claims.Name)
	assert.Equal(t, oauthAPI.IDTokenClaims.Email, claims.Email)
	assert.Equal(t, oauthAPI.IDTokenClaims.Picture, claims.Picture)
}

func TestService_NewAccessToken_Success(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	// Act
	token, err := service.NewAccessToken(1, "Jo Liao", "user@example.com", "https://example.com/picture.jpg")
	require.NoError(t, err)

	// Assert
	mockUserService.AssertExpectations(t)

	actual := &TokenClaims{}
	_, err = jwt.ParseWithClaims(token, actual, func(token *jwt.Token) (interface{}, error) {
		return config.AccessTokenSecret, nil
	})
	require.NoError(t, err)
	expected := &TokenClaims{
		Name:    "Jo Liao",
		Email:   "user@example.com",
		Picture: "https://example.com/picture.jpg",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  actual.IssuedAt,
			ExpiresAt: actual.ExpiresAt,
		},
	}
	assert.Equal(t, expected, actual)
	assert.WithinDuration(t, time.Now(), actual.IssuedAt.Time, time.Second)
	assert.WithinDuration(t, time.Now().Add(config.AccessTokenTTL), actual.ExpiresAt.Time, time.Second)
}

func TestService_NewRefreshToken_Success(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	// Act
	token, err := service.NewRefreshToken(1)
	require.NoError(t, err)

	// Assert
	mockUserService.AssertExpectations(t)

	actual := &TokenClaims{}
	_, err = jwt.ParseWithClaims(token, actual, func(token *jwt.Token) (interface{}, error) {
		return config.RefreshTokenSecret, nil
	})
	require.NoError(t, err)
	expected := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  actual.IssuedAt,
			ExpiresAt: actual.ExpiresAt,
		},
	}
	assert.Equal(t, expected, actual)
	assert.WithinDuration(t, time.Now(), actual.IssuedAt.Time, time.Second)
	assert.WithinDuration(t, time.Now().Add(config.RefreshTokenTTL), actual.ExpiresAt.Time, time.Second)
}

func TestService_ParseAccessToken_Success(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	expectedClaims := &TokenClaims{
		Email:   "user@example.com",
		Name:    "Jo Liao",
		Picture: "https://example.com/picture.jpg",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  jwt.NewNumericDate(time.Unix(1, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(10000000000, 0)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims).SignedString(config.AccessTokenSecret)
	require.NoError(t, err)

	// Act
	actualClaims, err := service.ParseAccessToken(token)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedClaims, actualClaims)
}
func TestService_ParseAccessToken_InvalidToken(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	expectedClaims := &TokenClaims{
		Email:   "user@example.com",
		Name:    "Jo Liao",
		Picture: "https://example.com/picture.jpg",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  jwt.NewNumericDate(time.Unix(1, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(10000000000, 0)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims).SignedString([]byte("invalid-secret"))
	require.NoError(t, err)

	// Act
	claims, err := service.ParseAccessToken(token)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claims)
}
func TestService_ParseAccessToken_InvalidIssuer(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	expectedClaims := &TokenClaims{
		Email:   "user@example.com",
		Name:    "Jo Liao",
		Picture: "https://example.com/picture.jpg",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "invalid-issuer",
			IssuedAt:  jwt.NewNumericDate(time.Unix(1, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(10000000000, 0)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims).SignedString(config.AccessTokenSecret)
	require.NoError(t, err)

	// Act
	claims, err := service.ParseAccessToken(token)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claims)
}
func TestService_ParseAccessToken_ExpiredToken(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	expectedClaims := &TokenClaims{
		Email:   "user@example.com",
		Name:    "Jo Liao",
		Picture: "https://example.com/picture.jpg",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  jwt.NewNumericDate(time.Unix(1, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(1, 0)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims).SignedString(config.AccessTokenSecret)
	require.NoError(t, err)

	// Act
	claims, err := service.ParseAccessToken(token)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claims)
}

func TestService_ParseRefreshToken_Success(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	expectedClaims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  jwt.NewNumericDate(time.Unix(1, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(10000000000, 0)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims).SignedString(config.RefreshTokenSecret)
	require.NoError(t, err)

	// Act
	actualClaims, err := service.ParseRefreshToken(token)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedClaims, actualClaims)
}
func TestService_ParseRefreshToken_InvalidToken(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	expectedClaims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  jwt.NewNumericDate(time.Unix(1, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(10000000000, 0)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims).SignedString([]byte("invalid-secret"))
	require.NoError(t, err)

	// Act
	claims, err := service.ParseRefreshToken(token)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claims)
}
func TestService_ParseRefreshToken_InvalidIssuer(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	expectedClaims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "invalid-issuer",
			IssuedAt:  jwt.NewNumericDate(time.Unix(1, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(10000000000, 0)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims).SignedString(config.RefreshTokenSecret)
	require.NoError(t, err)

	// Act
	claims, err := service.ParseRefreshToken(token)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claims)
}
func TestService_ParseRefreshToken_ExpiredToken(t *testing.T) {
	// Arrange
	mockUserService := &MockUserService{}
	config := NewMockConfig("")
	service := NewService(config, mockUserService)

	expectedClaims := &TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  jwt.NewNumericDate(time.Unix(1, 0)),
			ExpiresAt: jwt.NewNumericDate(time.Unix(1, 0)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expectedClaims).SignedString(config.RefreshTokenSecret)
	require.NoError(t, err)

	// Act
	claims, err := service.ParseRefreshToken(token)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claims)
}
