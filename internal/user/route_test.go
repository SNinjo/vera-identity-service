package user

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"vera-identity-service/internal/apperror"
	"vera-identity-service/internal/db"
	"vera-identity-service/internal/test"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func setupAuthConfig(oauthURL string) {
	Init(&AuthConfig{
		BaseURL:           "http://mock-base-url",
		FrontendURL:       "http://mock-frontend-url",
		OAuthClientID:     "mock-client-id",
		OAuthClientSecret: "mock-client-secret",
		OAuthEndpoint: oauth2.Endpoint{
			AuthURL:  oauthURL + "/auth",
			TokenURL: oauthURL + "/token",
		},
		AccessTokenSecret:  "mock-access-token-secret",
		AccessTokenTTL:     1 * time.Hour,
		RefreshTokenSecret: "mock-refresh-token-secret",
		RefreshTokenTTL:    2 * time.Hour,
	})
}
func setupOAuthServer(t *testing.T, expectedOAuthCode string) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		require.NoError(t, err)
		code := r.FormValue("code")
		if code != expectedOAuthCode {
			t.Errorf("unexpected code value: got %s, want %s", code, expectedOAuthCode)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{
			"sub": "1234567890",
			"name": "Jo Liao",
			"email": "user@example.com",
			"picture": "https://example.com/picture.png"
		}`))
		mockIDToken := header + "." + payload + "."
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{
			"id_token": "%s",
			"access_token": "mock_access_token",
			"token_type": "Bearer",
			"expires_in": 3600,
			"refresh_token": "mock_refresh_token"
		}`, mockIDToken)))
	})
	return httptest.NewServer(mux)
}

func TestAPI_AuthLogin_Success(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)

	req, err := http.NewRequest("GET", "/auth/login", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	u, err := url.Parse(location)
	require.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", u.Scheme+"://"+u.Host)
	assert.Equal(t, "/auth", u.Path)
	q := u.Query()
	assert.Equal(t, "mock-client-id", q.Get("client_id"))
	assert.Equal(t, "http://mock-base-url/auth/callback", q.Get("redirect_uri"))
	assert.Equal(t, "openid email profile", q.Get("scope"))
	assert.Equal(t, "code", q.Get("response_type"))
	assert.Equal(t, "offline", q.Get("access_type"))
	assert.Equal(t, "state", q.Get("state"))
}

func TestAPI_AuthCallback_Success(t *testing.T) {
	server := setupOAuthServer(t, "mock-code")
	defer server.Close()
	setupAuthConfig(server.URL)
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		Email: "user@example.com",
	})

	req, err := http.NewRequest("GET", "/auth/callback?code=mock-code&state=mock-state", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	u, err := url.Parse(location)
	require.NoError(t, err)
	assert.Equal(t, "http://mock-frontend-url", u.Scheme+"://"+u.Host)
	assert.Empty(t, u.Path)

	accessToken := u.Query().Get("access_token")
	accesstokenClaims := tokenClaims{}
	_, err = jwt.ParseWithClaims(accessToken, &accesstokenClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte("mock-access-token-secret"), nil
	})
	require.NoError(t, err)
	assert.Equal(t, "1", accesstokenClaims.Subject)
	assert.Equal(t, "Jo Liao", accesstokenClaims.Name)
	assert.Equal(t, "user@example.com", accesstokenClaims.Email)
	assert.Equal(t, "https://example.com/picture.png", accesstokenClaims.Picture)
	assert.Equal(t, "identity@vera.sninjo.com", accesstokenClaims.Issuer)
	assert.InDelta(t, time.Now().Unix(), accesstokenClaims.IssuedAt.Time.Unix(), 5)
	assert.InDelta(t, time.Now().Unix()+3600, accesstokenClaims.ExpiresAt.Time.Unix(), 5)

	cookies := w.Result().Cookies()
	var refreshTokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshTokenCookie = cookie
			break
		}
	}
	assert.NotNil(t, refreshTokenCookie)
	assert.True(t, refreshTokenCookie.HttpOnly)
	assert.True(t, refreshTokenCookie.Secure)
	refreshtokenClaims := tokenClaims{}
	_, err = jwt.ParseWithClaims(refreshTokenCookie.Value, &refreshtokenClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte("mock-refresh-token-secret"), nil
	})
	require.NoError(t, err)
	assert.Equal(t, "1", refreshtokenClaims.Subject)
	assert.InDelta(t, time.Now().Unix(), refreshtokenClaims.IssuedAt.Time.Unix(), 5)
	assert.InDelta(t, time.Now().Unix()+7200, refreshtokenClaims.ExpiresAt.Time.Unix(), 5)
}

func TestAPI_AuthRefresh_Success(t *testing.T) {
	server := setupOAuthServer(t, "mock-code")
	defer server.Close()
	setupAuthConfig(server.URL)
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		ID:      1,
		Email:   "user@example.com",
		Name:    test.StringPtr("Jo Liao"),
		Picture: "https://example.com/picture.png",
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "1",
		"iss": "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-refresh-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/auth/refresh", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var resp tokenResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	accesstokenClaims := tokenClaims{}
	_, err = jwt.ParseWithClaims(resp.AccessToken, &accesstokenClaims, func(token *jwt.Token) (interface{}, error) {
		return []byte("mock-access-token-secret"), nil
	})
	require.NoError(t, err)
	assert.Equal(t, "1", accesstokenClaims.Subject)
	assert.Equal(t, "Jo Liao", accesstokenClaims.Name)
	assert.Equal(t, "user@example.com", accesstokenClaims.Email)
	assert.Equal(t, "https://example.com/picture.png", accesstokenClaims.Picture)
	assert.Equal(t, "identity@vera.sninjo.com", accesstokenClaims.Issuer)
	assert.InDelta(t, time.Now().Unix(), accesstokenClaims.IssuedAt.Time.Unix(), 5)
	assert.InDelta(t, time.Now().Unix()+3600, accesstokenClaims.ExpiresAt.Time.Unix(), 5)
}

func TestAPI_AuthVerify_Success(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		ID:           1,
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user1@example.com",
		Picture:      "https://example.com/picture1.jpg",
		LastLoginSub: &[]string{"mock-sub-1"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"name":    "Jo Liao",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-access-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/auth/verify", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
}
func TestAPI_AuthVerify_MissingAuthHeader(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("POST", "/auth/verify", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_008", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}
func TestAPI_AuthVerify_InvalidAccessToken(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("POST", "/auth/verify", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_006", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}

func TestAPI_UsersList_Success(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		ID:           1,
		Name:         test.StringPtr("Jo Liao 1"),
		Email:        "user1@example.com",
		Picture:      "https://example.com/picture1.jpg",
		LastLoginSub: &[]string{"mock-sub-1"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})
	db.DB.Create(&User{
		ID:           2,
		Name:         test.StringPtr("Jo Liao 2"),
		Email:        "user2@example.com",
		Picture:      "https://example.com/picture2.jpg",
		LastLoginSub: &[]string{"mock-sub-2"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(2, 0)}[0],
		CreatedAt:    time.Unix(2, 0),
		UpdatedAt:    time.Unix(2, 0),
	})
	deleted := &User{
		ID:           3,
		Name:         test.StringPtr("Jo Liao 3"),
		Email:        "user3@example.com",
		Picture:      "https://example.com/picture3.jpg",
		LastLoginSub: &[]string{"mock-sub-3"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(3, 0)}[0],
		CreatedAt:    time.Unix(3, 0),
		UpdatedAt:    time.Unix(3, 0),
	}
	db.DB.Create(deleted)
	db.DB.Delete(deleted)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-access-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var userResponses []userResponse
	err = json.Unmarshal(w.Body.Bytes(), &userResponses)
	require.NoError(t, err)
	assert.Len(t, userResponses, 2)
	assert.Equal(t, 1, userResponses[0].ID)
	assert.Equal(t, "Jo Liao 1", *userResponses[0].Name)
	assert.Equal(t, "user1@example.com", userResponses[0].Email)
	assert.Equal(t, "https://example.com/picture1.jpg", userResponses[0].Picture)
	assert.Equal(t, time.Unix(1, 0), *userResponses[0].LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), userResponses[0].CreatedAt)
	assert.Equal(t, time.Unix(1, 0), userResponses[0].UpdatedAt)
	assert.Equal(t, 2, userResponses[1].ID)
	assert.Equal(t, "Jo Liao 2", *userResponses[1].Name)
	assert.Equal(t, "user2@example.com", userResponses[1].Email)
	assert.Equal(t, "https://example.com/picture2.jpg", userResponses[1].Picture)
	assert.Equal(t, time.Unix(2, 0), *userResponses[1].LastLoginAt)
	assert.Equal(t, time.Unix(2, 0), userResponses[1].CreatedAt)
	assert.Equal(t, time.Unix(2, 0), userResponses[1].UpdatedAt)
}
func TestAPI_UsersList_MissingAuthHeader(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_008", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}
func TestAPI_UsersList_InvalidAccessToken(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_006", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}

func TestAPI_UsersCreate_Success(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"name":    "Jo Liao",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-access-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users", strings.NewReader(`{"email": "newuser@example.com"}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var user userResponse
	err = json.Unmarshal(w.Body.Bytes(), &user)
	require.NoError(t, err)
	assert.Equal(t, 2, user.ID)
	assert.Equal(t, "newuser@example.com", user.Email)
	assert.Equal(t, "", user.Picture)
	assert.Nil(t, user.LastLoginAt)
	assert.InDelta(t, time.Now().Unix(), user.CreatedAt.Unix(), 5)
	assert.InDelta(t, time.Now().Unix(), user.UpdatedAt.Unix(), 5)

	req, err = http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var users []userResponse
	err = json.Unmarshal(w.Body.Bytes(), &users)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, "Jo Liao", *users[0].Name)
	assert.Equal(t, "user@example.com", users[0].Email)
	assert.Equal(t, "https://example.com/picture.jpg", users[0].Picture)
	assert.Equal(t, time.Unix(1, 0), *users[0].LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), users[0].CreatedAt)
	assert.Equal(t, time.Unix(1, 0), users[0].UpdatedAt)
	assert.Equal(t, 2, users[1].ID)
	assert.Nil(t, users[1].Name)
	assert.Equal(t, "newuser@example.com", users[1].Email)
	assert.Equal(t, "", users[1].Picture)
	assert.Nil(t, users[1].LastLoginAt)
	assert.InDelta(t, time.Now().Unix(), users[1].CreatedAt.Unix(), 5)
	assert.InDelta(t, time.Now().Unix(), users[1].UpdatedAt.Unix(), 5)
}
func TestAPI_UsersCreate_MissingAuthHeader(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("POST", "/users", strings.NewReader(`{"email": "newuser@example.com"}`))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_008", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}
func TestAPI_UsersCreate_InvalidAccessToken(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("POST", "/users", strings.NewReader(`{"email": "newuser@example.com"}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_006", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}
func TestAPI_UsersCreate_ErrorEmailFormat(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"name":    "Jo Liao",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-access-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users", strings.NewReader(`{"email": "invalid-email"}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPI_UsersUpdate_Success(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"name":    "Jo Liao",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-access-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("PATCH", "/users/1", strings.NewReader(`{"email": "newuser@example.com"}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var user userResponse
	err = json.Unmarshal(w.Body.Bytes(), &user)
	require.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "Jo Liao", *user.Name)
	assert.Equal(t, "newuser@example.com", user.Email)
	assert.Equal(t, "https://example.com/picture.jpg", user.Picture)
	assert.Equal(t, time.Unix(1, 0), *user.LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), user.CreatedAt)
	assert.InDelta(t, time.Now().Unix(), user.UpdatedAt.Unix(), 5)

	req, err = http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var users []userResponse
	err = json.Unmarshal(w.Body.Bytes(), &users)
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, "Jo Liao", *users[0].Name)
	assert.Equal(t, "newuser@example.com", users[0].Email)
	assert.Equal(t, "https://example.com/picture.jpg", users[0].Picture)
	assert.Equal(t, time.Unix(1, 0), *users[0].LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), users[0].CreatedAt)
	assert.InDelta(t, time.Now().Unix(), users[0].UpdatedAt.Unix(), 5)
}
func TestAPI_UsersUpdate_MissingAuthHeader(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("PATCH", "/users/1", strings.NewReader(`{"email": "newuser@example.com"}`))
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_008", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}
func TestAPI_UsersUpdate_InvalidAccessToken(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("PATCH", "/users/1", strings.NewReader(`{"email": "newuser@example.com"}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_006", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}
func TestAPI_UsersUpdate_ErrorIDFormat(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-access-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("PATCH", "/users/invalid-id", strings.NewReader(`{"email": "newuser@example.com"}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}
func TestAPI_UsersUpdate_ErrorEmailFormat(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-access-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("PATCH", "/users/1", strings.NewReader(`{"email": "invalid-email"}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPI_UsersDelete_Success(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})
	db.DB.Create(&User{
		Name:         test.StringPtr("Jo Liao 2"),
		Email:        "user2@example.com",
		Picture:      "https://example.com/picture2.jpg",
		LastLoginSub: &[]string{"mock-sub2"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(2, 0)}[0],
		CreatedAt:    time.Unix(2, 0),
		UpdatedAt:    time.Unix(2, 0),
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-access-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("DELETE", "/users/2", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "", w.Body.String())

	req, err = http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var users []userResponse
	err = json.Unmarshal(w.Body.Bytes(), &users)
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, "Jo Liao", *users[0].Name)
	assert.Equal(t, "user@example.com", users[0].Email)
	assert.Equal(t, "https://example.com/picture.jpg", users[0].Picture)
	assert.Equal(t, time.Unix(1, 0), *users[0].LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), users[0].CreatedAt)
	assert.Equal(t, time.Unix(1, 0), users[0].UpdatedAt)
}
func TestAPI_UsersDelete_MissingAuthHeader(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("DELETE", "/users/1", nil)
	require.NoError(t, err)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_008", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}
func TestAPI_UsersDelete_InvalidAccessToken(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	req, err := http.NewRequest("DELETE", "/users/1", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp apperror.Response
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "401_01_006", resp.Code)
	assert.InDelta(t, time.Now().Unix(), resp.Timestamp.Unix(), 5)
}
func TestAPI_UsersDelete_ErrorIDFormat(t *testing.T) {
	setupAuthConfig("http://localhost:8080")
	test.SetupLogger()
	router := test.SetupRouter(RegisterRoutes)
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"name":    "Jo Liao",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
	}).SignedString([]byte("mock-access-token-secret"))
	require.NoError(t, err)

	req, err := http.NewRequest("DELETE", "/users/invalid-id", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
}
