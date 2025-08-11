package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"
	"vera-identity-service/internal/app"
	"vera-identity-service/internal/auth"
	"vera-identity-service/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var a *app.App

func TestMain(m *testing.M) {
	// Setup
	gin.SetMode(gin.TestMode)

	dbURL, closeDB, err := SetupPostgresql()
	if err != nil {
		log.Fatal(err)
	}

	envs := map[string]string{
		"BASE_URL":     "http://mock-base-url",
		"DATABASE_URL": dbURL,
		"SITE_URL":     "http://mock-site-url",

		"GOOGLE_CLIENT_ID":     "mock-google-client-id",
		"GOOGLE_CLIENT_SECRET": "mock-google-client-secret",

		"ACCESS_TOKEN_TTL":     "1h",
		"ACCESS_TOKEN_SECRET":  "mock-access-token-secret",
		"REFRESH_TOKEN_TTL":    "2h",
		"REFRESH_TOKEN_SECRET": "mock-refresh-token-secret",
	}
	for key, value := range envs {
		err = os.Setenv(key, value)
		if err != nil {
			log.Fatal(err)
		}
	}

	a, err = app.InitApp()
	if err != nil {
		log.Fatal(err)
	}

	err = a.DB.AutoMigrate(&user.User{})
	if err != nil {
		log.Fatal(err)
	}

	// Run
	code := m.Run()

	// Teardown
	a.Close()
	closeDB()

	os.Exit(code)
}

func createTestRequest(method, path string, body interface{}, token string) (*http.Request, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	return req, nil
}

func TestAPI_AuthLogin_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("GET", "/auth/login", nil, "")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusFound, w.Code)

	actualURL, err := url.Parse(w.Header().Get("Location"))
	require.NoError(t, err)
	expectedURL, err := url.Parse(
		google.Endpoint.AuthURL +
			"?access_type=offline" +
			"&client_id=mock-google-client-id" +
			"&redirect_uri=http%3A%2F%2Fmock-base-url%2Fauth%2Fcallback" +
			"&response_type=code" +
			"&scope=openid+email+profile" +
			"&state=state",
	)
	require.NoError(t, err)
	assert.Equal(t, expectedURL, actualURL)
}

func TestAPI_AuthCallback_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)
	oauthAPI := SetupOAuthAPI()
	a.Config.OAuth2.Endpoint = oauth2.Endpoint{
		TokenURL: oauthAPI.URL + "/token",
	}

	user := user.User{
		ID:    1,
		Email: oauthAPI.IDTokenClaims.Email,
	}
	err = a.DB.Create(&user).Error
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("GET", "/auth/callback?code="+oauthAPI.AuthorizationCode, nil, "")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusFound, w.Code)

	location := w.Header().Get("Location")
	u, err := url.Parse(location)
	require.NoError(t, err)
	assert.Equal(t, "http://mock-site-url", u.Scheme+"://"+u.Host)

	accessToken := u.Query().Get("access_token")
	require.NotEmpty(t, accessToken)
	actualClaims, err := a.AuthService.ParseAccessToken(accessToken)
	require.NoError(t, err)
	expectedClaims := &auth.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(user.ID),
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  actualClaims.IssuedAt,
			ExpiresAt: actualClaims.ExpiresAt,
		},
		Name:    oauthAPI.IDTokenClaims.Name,
		Email:   oauthAPI.IDTokenClaims.Email,
		Picture: oauthAPI.IDTokenClaims.Picture,
	}
	assert.Equal(t, expectedClaims, actualClaims)
	assert.WithinDuration(t, time.Now(), actualClaims.IssuedAt.Time, time.Second)
	assert.WithinDuration(t, time.Now().Add(a.Config.AccessTokenTTL), actualClaims.ExpiresAt.Time, time.Second)

	cookies := w.Result().Cookies()
	var refreshTokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshTokenCookie = cookie
			break
		}
	}
	assert.Equal(t, "/", refreshTokenCookie.Path)
	assert.Empty(t, refreshTokenCookie.Domain)
	assert.Equal(t, int(a.Config.RefreshTokenTTL.Seconds()), refreshTokenCookie.MaxAge)
	assert.True(t, refreshTokenCookie.HttpOnly)
	assert.True(t, refreshTokenCookie.Secure)
	actualClaims, err = a.AuthService.ParseRefreshToken(refreshTokenCookie.Value)
	require.NoError(t, err)
	expectedClaims = &auth.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(user.ID),
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  actualClaims.IssuedAt,
			ExpiresAt: actualClaims.ExpiresAt,
		},
	}
	assert.Equal(t, expectedClaims, actualClaims)
	assert.WithinDuration(t, time.Now(), actualClaims.IssuedAt.Time, time.Second)
	assert.WithinDuration(t, time.Now().Add(a.Config.RefreshTokenTTL), actualClaims.ExpiresAt.Time, time.Second)
}

func TestAPI_AuthRefresh_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	user := user.User{
		ID:      1,
		Name:    StringPtr("name"),
		Email:   "user@example.com",
		Picture: StringPtr("https://example.com/picture.jpg"),
	}
	err = a.DB.Create(&user).Error
	require.NoError(t, err)
	refreshToken, err := a.AuthService.NewRefreshToken(user.ID)
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("POST", "/auth/refresh", nil, "")
	req.AddCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   "",
		MaxAge:   int(a.Config.RefreshTokenTTL.Seconds()),
		HttpOnly: true,
		Secure:   true,
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusOK, w.Code)

	var resp auth.TokenResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	actualClaims, err := a.AuthService.ParseAccessToken(resp.AccessToken)
	require.NoError(t, err)
	expectedClaims := &auth.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(user.ID),
			Issuer:    "identity@vera.sninjo.com",
			IssuedAt:  actualClaims.IssuedAt,
			ExpiresAt: actualClaims.ExpiresAt,
		},
		Name:    *user.Name,
		Email:   user.Email,
		Picture: *user.Picture,
	}
	assert.Equal(t, expectedClaims, actualClaims)
	assert.WithinDuration(t, time.Now(), actualClaims.IssuedAt.Time, time.Second)
	assert.WithinDuration(t, time.Now().Add(a.Config.AccessTokenTTL), actualClaims.ExpiresAt.Time, time.Second)
}

func TestAPI_AuthVerify_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	accessToken, err := a.AuthService.NewAccessToken(1, "", "", "")
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("POST", "/auth/verify", nil, accessToken)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)
}

func TestAPI_UsersGet_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	user1 := user.User{
		ID:           1,
		Name:         StringPtr("name1"),
		Email:        "user1@example.com",
		Picture:      StringPtr("https://example.com/picture1.jpg"),
		LastLoginSub: StringPtr("sub1"),
		LastLoginAt:  nil,
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	}
	user2 := user.User{
		ID:           2,
		Name:         StringPtr("name2"),
		Email:        "user2@example.com",
		Picture:      StringPtr("https://example.com/picture2.jpg"),
		LastLoginSub: StringPtr("sub2"),
		LastLoginAt:  TimePtr(time.Unix(2, 0)),
		CreatedAt:    time.Unix(2, 0),
		UpdatedAt:    time.Unix(2, 0),
	}
	err = a.DB.Create(&user1).Error
	require.NoError(t, err)
	err = a.DB.Create(&user2).Error
	require.NoError(t, err)
	accessToken, err := a.AuthService.NewAccessToken(1, "", "", "")
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("GET", "/users", nil, accessToken)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	fmt.Println(w.Body.String())
	require.Equal(t, http.StatusOK, w.Code)

	var actual []user.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &actual)
	require.NoError(t, err)
	expected := []user.UserResponse{
		{
			ID:          user1.ID,
			Name:        user1.Name,
			Email:       user1.Email,
			Picture:     user1.Picture,
			LastLoginAt: nil,
			CreatedAt:   user1.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user1.UpdatedAt.Format(time.RFC3339),
		},
		{
			ID:          user2.ID,
			Name:        user2.Name,
			Email:       user2.Email,
			Picture:     user2.Picture,
			LastLoginAt: StringPtr(user2.LastLoginAt.Format(time.RFC3339)),
			CreatedAt:   user2.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user2.UpdatedAt.Format(time.RFC3339),
		},
	}
	assert.Equal(t, expected, actual)
}

func TestAPI_UsersPost_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	accessToken, err := a.AuthService.NewAccessToken(1, "", "", "")
	require.NoError(t, err)

	// Act
	body := user.RequestBody{
		Email: "new@example.com",
	}
	req, err := createTestRequest("POST", "/users", body, accessToken)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)

	req, err = createTestRequest("GET", "/users", nil, accessToken)
	require.NoError(t, err)
	w = httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	var resp []user.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	actual := resp[0]
	expected := user.UserResponse{
		ID:          1,
		Name:        nil,
		Email:       body.Email,
		Picture:     nil,
		LastLoginAt: nil,
		CreatedAt:   actual.CreatedAt,
		UpdatedAt:   actual.UpdatedAt,
	}
	assert.Equal(t, expected, actual)
	actualCreatedAt, err := time.Parse(time.RFC3339, actual.CreatedAt)
	require.NoError(t, err)
	actualUpdatedAt, err := time.Parse(time.RFC3339, actual.UpdatedAt)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), actualCreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), actualUpdatedAt, time.Second)
}

func TestAPI_UsersPatch_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	existingUser := user.User{
		ID:           1,
		Name:         StringPtr("name1"),
		Email:        "user1@example.com",
		Picture:      StringPtr("https://example.com/picture1.jpg"),
		LastLoginSub: StringPtr("sub1"),
		LastLoginAt:  nil,
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	}
	err = a.DB.Create(&existingUser).Error
	require.NoError(t, err)
	accessToken, err := a.AuthService.NewAccessToken(1, "", "", "")
	require.NoError(t, err)

	// Act
	body := user.RequestBody{
		Email: "new@example.com",
	}
	req, err := createTestRequest("PATCH", "/users/1", body, accessToken)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)

	req, err = createTestRequest("GET", "/users", nil, accessToken)
	require.NoError(t, err)
	w = httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	var resp []user.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	actual := resp[0]
	expected := user.UserResponse{
		ID:          existingUser.ID,
		Name:        existingUser.Name,
		Email:       body.Email,
		Picture:     existingUser.Picture,
		LastLoginAt: nil,
		CreatedAt:   existingUser.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   actual.UpdatedAt,
	}
	assert.Equal(t, expected, actual)
	actualUpdatedAt, err := time.Parse(time.RFC3339, actual.UpdatedAt)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), actualUpdatedAt, time.Second)
}

func TestAPI_UsersDelete_Success(t *testing.T) {
	// Arrange
	err := CleanupTables(a.DB)
	require.NoError(t, err)

	existingUser := user.User{
		ID:           1,
		Name:         StringPtr("name1"),
		Email:        "user1@example.com",
		Picture:      StringPtr("https://example.com/picture1.jpg"),
		LastLoginSub: StringPtr("sub1"),
		LastLoginAt:  nil,
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	}
	err = a.DB.Create(&existingUser).Error
	require.NoError(t, err)
	accessToken, err := a.AuthService.NewAccessToken(1, "", "", "")
	require.NoError(t, err)

	// Act
	req, err := createTestRequest("DELETE", "/users/1", nil, accessToken)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	// Assert
	require.Equal(t, http.StatusNoContent, w.Code)

	req, err = createTestRequest("GET", "/users", nil, accessToken)
	require.NoError(t, err)
	w = httptest.NewRecorder()
	a.Router.ServeHTTP(w, req)

	var resp []user.UserResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Empty(t, resp)
}

func TestAPI_AllURLs_Unauthorized(t *testing.T) {
	tests := []struct {
		method string
		path   string
	}{
		{"POST", "/auth/verify"},
		{"GET", "/users"},
		{"POST", "/users"},
		{"PATCH", "/users/1"},
		{"DELETE", "/users/1"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			// Arrange
			err := CleanupTables(a.DB)
			require.NoError(t, err)

			token, err := jwt.New(jwt.SigningMethodHS256).SignedString([]byte("invalid-token-secret"))
			require.NoError(t, err)

			// Act
			req, err := createTestRequest(tt.method, tt.path, nil, token)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			a.Router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "invalid access token")
		})
	}
}
