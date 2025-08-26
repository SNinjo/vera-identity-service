package auth

import (
	"time"

	"github.com/sninjo/vera-identity-service/internal/config"
	"github.com/sninjo/vera-identity-service/internal/user"

	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

func NewMockConfig(oauthURL string) *config.Config {
	return &config.Config{
		BaseURL: "http://mock-base-url",
		SiteURL: "http://mock-site-url",

		OAuth2: &oauth2.Config{
			ClientID:     "mock-client-id",
			ClientSecret: "mock-client-secret",
			RedirectURL:  "http://mock-base-url/auth/callback",
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  oauthURL + "/auth",
				TokenURL: oauthURL + "/token",
			},
		},

		AccessTokenSecret:  []byte("mock-access-token-secret"),
		RefreshTokenSecret: []byte("mock-refresh-token-secret"),
		AccessTokenTTL:     time.Hour,
		RefreshTokenTTL:    2 * time.Hour,
	}
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(email string) error {
	args := m.Called(email)
	return args.Error(0)
}
func (m *MockUserService) GetUserByID(id int) (*user.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}
func (m *MockUserService) GetUserByEmail(email string) (*user.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}
func (m *MockUserService) GetUsers() ([]user.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]user.User), args.Error(1)
}
func (m *MockUserService) UpdateUser(id int, email string) error {
	args := m.Called(id, email)
	return args.Error(0)
}
func (m *MockUserService) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockUserService) RecordUserLogin(id int, name, picture, loginSub string) error {
	args := m.Called(id, name, picture, loginSub)
	return args.Error(0)
}
