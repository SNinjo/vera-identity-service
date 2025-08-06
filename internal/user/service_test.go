package user

import (
	"testing"
	"time"
	"vera-identity-service/internal/apperror"
	"vera-identity-service/internal/db"
	"vera-identity-service/internal/test"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUnit_NewJWT_Normal(t *testing.T) {
	token, err := newJWT(
		&User{ID: 1, Name: test.StringPtr("Jo Liao"), Email: "user@example.com", Picture: "https://example.com/picture.jpg"},
		"mock-secret",
		1*time.Hour,
	)
	require.NoError(t, err)

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("mock-secret"), nil
	})
	require.NoError(t, err)
	assert.Equal(t, "1", claims["sub"])
	assert.Equal(t, "Jo Liao", claims["name"])
	assert.Equal(t, "user@example.com", claims["email"])
	assert.Equal(t, "https://example.com/picture.jpg", claims["picture"])
	assert.Equal(t, "identity@vera.sninjo.com", claims["iss"])
	assert.InDelta(t, time.Now().Unix(), claims["iat"].(float64), 5)
	assert.InDelta(t, time.Now().Unix()+3600, claims["exp"].(float64), 5)
}
func TestUnit_NewJWT_OnlyID(t *testing.T) {
	token, err := newJWT(
		&User{ID: 1},
		"mock-secret",
		1*time.Hour,
	)
	require.NoError(t, err)

	claims := jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("mock-secret"), nil
	})
	require.NoError(t, err)
	assert.Equal(t, "1", claims["sub"])
	assert.Nil(t, claims["email"])
	assert.Nil(t, claims["picture"])
	assert.Equal(t, "identity@vera.sninjo.com", claims["iss"])
	assert.InDelta(t, time.Now().Unix(), claims["iat"].(float64), 5)
	assert.InDelta(t, time.Now().Unix()+3600, claims["exp"].(float64), 5)
}

func TestUnit_ParseJWT_Normal(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"email":   "user@example.com",
		"name":    "John Doe",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
		"iat":     time.Unix(1, 0).Unix(),
		"exp":     time.Unix(10000000000, 0).Unix(),
	}).SignedString([]byte("mock-secret"))
	require.NoError(t, err)

	claims, err := parseJWT(token, "mock-secret")
	require.NoError(t, err)
	assert.Equal(t, "1", claims.Subject)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, "John Doe", claims.Name)
	assert.Equal(t, "https://example.com/picture.jpg", claims.Picture)
	assert.Equal(t, "identity@vera.sninjo.com", claims.Issuer)
	assert.Equal(t, time.Unix(1, 0), claims.IssuedAt.Time)
	assert.Equal(t, time.Unix(10000000000, 0), claims.ExpiresAt.Time)
}
func TestUnit_ParseJWT_InvalidToken(t *testing.T) {
	claims, err := parseJWT("invalid-token", "only-for-test")
	assert.Error(t, err)
	assert.Nil(t, claims)
}
func TestUnit_ParseJWT_InvalidSecret(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
		"iat":     time.Unix(1, 0).Unix(),
		"exp":     time.Unix(10000000000, 0).Unix(),
	}).SignedString([]byte("mock-secret"))
	require.NoError(t, err)

	claims, err := parseJWT(token, "invalid-secret")
	assert.Error(t, err)
	assert.Nil(t, claims)
}
func TestUnit_ParseJWT_InvalidIssuer(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "invalid-issuer",
		"iat":     time.Unix(1, 0).Unix(),
		"exp":     time.Unix(10000000000, 0).Unix(),
	}).SignedString([]byte("mock-secret"))
	require.NoError(t, err)

	claims, err := parseJWT(token, "mock-secret")
	assert.Equal(t, apperror.CodeInvalidTokenIssuer, err.(*apperror.AppError).Code)
	assert.Nil(t, claims)
}
func TestUnit_ParseJWT_ExpiredToken(t *testing.T) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":     "1",
		"email":   "user@example.com",
		"picture": "https://example.com/picture.jpg",
		"iss":     "identity@vera.sninjo.com",
		"iat":     time.Unix(1, 0).Unix(),
		"exp":     time.Unix(1, 0).Unix(),
	}).SignedString([]byte("mock-secret"))
	require.NoError(t, err)

	claims, err := parseJWT(token, "mock-secret")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestUnit_GetUserByID_Normal(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		ID:           1,
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
		DeletedAt:    gorm.DeletedAt{},
	})

	u, err := getUserByID(1)
	require.NoError(t, err)
	assert.Equal(t, 1, u.ID)
	assert.Equal(t, "Jo Liao", *u.Name)
	assert.Equal(t, "user@example.com", u.Email)
	assert.Equal(t, "https://example.com/picture.jpg", u.Picture)
	assert.Equal(t, "mock-sub", *u.LastLoginSub)
	assert.Equal(t, time.Unix(1, 0), *u.LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), u.CreatedAt)
	assert.Equal(t, time.Unix(1, 0), u.UpdatedAt)
	assert.Equal(t, gorm.DeletedAt{}, u.DeletedAt)
}
func TestUnit_GetUserByID_NotFound(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	u, err := getUserByID(1)
	assert.NoError(t, err)
	assert.Nil(t, u)
}
func TestUnit_GetUserByID_Deleted(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	u := &User{
		ID:           1,
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	}
	db.DB.Create(u)
	db.DB.Delete(u)

	u, err := getUserByID(1)
	assert.NoError(t, err)
	assert.Nil(t, u)
}

func TestUnit_GetUserByEmail_Normal(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		ID:           1,
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
		DeletedAt:    gorm.DeletedAt{},
	})

	u, err := getUserByEmail("user@example.com")
	require.NoError(t, err)
	assert.Equal(t, 1, u.ID)
	assert.Equal(t, "Jo Liao", *u.Name)
	assert.Equal(t, "user@example.com", u.Email)
	assert.Equal(t, "https://example.com/picture.jpg", u.Picture)
	assert.Equal(t, "mock-sub", *u.LastLoginSub)
	assert.Equal(t, time.Unix(1, 0), *u.LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), u.CreatedAt)
	assert.Equal(t, time.Unix(1, 0), u.UpdatedAt)
	assert.Equal(t, gorm.DeletedAt{}, u.DeletedAt)
}
func TestUnit_GetUserByEmail_NotFound(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	u, err := getUserByEmail("user@example.com")
	assert.NoError(t, err)
	assert.Nil(t, u)
}
func TestUnit_GetUserByEmail_Deleted(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	u := &User{
		ID:           1,
		Name:         test.StringPtr("Jo Liao"),
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	}
	db.DB.Create(u)
	db.DB.Delete(u)

	u, err := getUserByEmail("user@example.com")
	assert.NoError(t, err)
	assert.Nil(t, u)
}

func TestUnit_GetUsers_Normal(t *testing.T) {
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
		DeletedAt:    gorm.DeletedAt{},
	})
	db.DB.Create(&User{
		ID:           2,
		Name:         test.StringPtr("Jo Liao 2"),
		Email:        "user2@example.com",
		Picture:      "https://example.com/picture2.jpg",
		LastLoginSub: &[]string{"mock-sub2"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(2, 0)}[0],
		CreatedAt:    time.Unix(2, 0),
		UpdatedAt:    time.Unix(2, 0),
		DeletedAt:    gorm.DeletedAt{},
	})
	db.DB.Create(&User{
		ID:           3,
		Name:         test.StringPtr("Jo Liao 3"),
		Email:        "user3@example.com",
		Picture:      "https://example.com/picture3.jpg",
		LastLoginSub: &[]string{"mock-sub3"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(3, 0)}[0],
		CreatedAt:    time.Unix(3, 0),
		UpdatedAt:    time.Unix(3, 0),
		DeletedAt:    gorm.DeletedAt{Time: time.Unix(3, 0)},
	})

	users, err := getUsers()
	require.NoError(t, err)
	assert.Equal(t, 3, len(users))
	assert.Equal(t, 1, users[0].ID)
	assert.Equal(t, "Jo Liao 1", *users[0].Name)
	assert.Equal(t, "user1@example.com", users[0].Email)
	assert.Equal(t, "https://example.com/picture1.jpg", users[0].Picture)
	assert.Equal(t, "mock-sub-1", *users[0].LastLoginSub)
	assert.Equal(t, time.Unix(1, 0), *users[0].LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), users[0].CreatedAt)
	assert.Equal(t, time.Unix(1, 0), users[0].UpdatedAt)
	assert.Equal(t, gorm.DeletedAt{}, users[0].DeletedAt)
	assert.Equal(t, 2, users[1].ID)
	assert.Equal(t, "Jo Liao 2", *users[1].Name)
	assert.Equal(t, "user2@example.com", users[1].Email)
	assert.Equal(t, "https://example.com/picture2.jpg", users[1].Picture)
	assert.Equal(t, "mock-sub2", *users[1].LastLoginSub)
	assert.Equal(t, time.Unix(2, 0), *users[1].LastLoginAt)
	assert.Equal(t, time.Unix(2, 0), users[1].CreatedAt)
	assert.Equal(t, time.Unix(2, 0), users[1].UpdatedAt)
	assert.Equal(t, gorm.DeletedAt{}, users[1].DeletedAt)
}

func TestUnit_CreateUser_Normal(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	user := &User{
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
	}
	err := createUser(user)
	require.NoError(t, err)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "user@example.com", user.Email)
	assert.Equal(t, "https://example.com/picture.jpg", user.Picture)
	assert.Equal(t, "mock-sub", *user.LastLoginSub)
	assert.Equal(t, time.Unix(1, 0), *user.LastLoginAt)
	assert.InDelta(t, time.Now().Unix(), user.CreatedAt.Unix(), 5)
	assert.InDelta(t, time.Now().Unix(), user.UpdatedAt.Unix(), 5)
	assert.Equal(t, gorm.DeletedAt{}, user.DeletedAt)

	var u User
	db.DB.Where("email = ?", "user@example.com").First(&u)
	assert.Equal(t, 1, u.ID)
	assert.Equal(t, "user@example.com", u.Email)
	assert.Equal(t, "https://example.com/picture.jpg", u.Picture)
	assert.Equal(t, "mock-sub", *u.LastLoginSub)
	assert.Equal(t, time.Unix(1, 0), *u.LastLoginAt)
	assert.InDelta(t, time.Now().Unix(), u.CreatedAt.Unix(), 5)
	assert.InDelta(t, time.Now().Unix(), u.UpdatedAt.Unix(), 5)
	assert.Equal(t, gorm.DeletedAt{}, u.DeletedAt)
}

func TestUnit_UpdateUser_Normal(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		ID:           1,
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
		DeletedAt:    gorm.DeletedAt{},
	})

	newUser, err := updateUser(1, &User{
		Email:        "user2@example.com",
		Picture:      "https://example.com/picture2.jpg",
		LastLoginSub: &[]string{"mock-sub2"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(2, 0)}[0],
	})
	require.NoError(t, err)
	assert.Equal(t, 1, newUser.ID)
	assert.Equal(t, "user2@example.com", newUser.Email)
	assert.Equal(t, "https://example.com/picture2.jpg", newUser.Picture)
	assert.Equal(t, "mock-sub2", *newUser.LastLoginSub)
	assert.Equal(t, time.Unix(2, 0), *newUser.LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), newUser.CreatedAt)
	assert.InDelta(t, time.Now().Unix(), newUser.UpdatedAt.Unix(), 5)
	assert.Equal(t, gorm.DeletedAt{}, newUser.DeletedAt)

	var u User
	db.DB.Where("id = ?", 1).First(&u)
	assert.Equal(t, 1, u.ID)
	assert.Equal(t, "user2@example.com", u.Email)
	assert.Equal(t, "https://example.com/picture2.jpg", u.Picture)
	assert.Equal(t, "mock-sub2", *u.LastLoginSub)
	assert.Equal(t, time.Unix(2, 0), *u.LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), u.CreatedAt)
	assert.InDelta(t, time.Now().Unix(), u.UpdatedAt.Unix(), 5)
	assert.Equal(t, gorm.DeletedAt{}, u.DeletedAt)
}
func TestUnit_UpdateUser_OnlyEmail(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		ID:           1,
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
		DeletedAt:    gorm.DeletedAt{},
	})

	newUser, err := updateUser(1, &User{
		Email: "user2@example.com",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, newUser.ID)
	assert.Equal(t, "user2@example.com", newUser.Email)
	assert.Equal(t, "https://example.com/picture.jpg", newUser.Picture)
	assert.Equal(t, "mock-sub", *newUser.LastLoginSub)
	assert.Equal(t, time.Unix(1, 0), *newUser.LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), newUser.CreatedAt)
	assert.InDelta(t, time.Now().Unix(), newUser.UpdatedAt.Unix(), 5)
	assert.Equal(t, gorm.DeletedAt{}, newUser.DeletedAt)

	var u User
	db.DB.Where("id = ?", 1).First(&u)
	assert.Equal(t, 1, u.ID)
	assert.Equal(t, "user2@example.com", u.Email)
	assert.Equal(t, "https://example.com/picture.jpg", u.Picture)
	assert.Equal(t, "mock-sub", *u.LastLoginSub)
	assert.Equal(t, time.Unix(1, 0), *u.LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), u.CreatedAt)
	assert.InDelta(t, time.Now().Unix(), u.UpdatedAt.Unix(), 5)
	assert.Equal(t, gorm.DeletedAt{}, u.DeletedAt)
}
func TestUnit_UpdateUser_NotFound(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	newUser, err := updateUser(1, &User{
		Email:        "user2@example.com",
		Picture:      "https://example.com/picture2.jpg",
		LastLoginSub: &[]string{"mock-sub2"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(2, 0)}[0],
	})
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeUserNotFound, err.(*apperror.AppError).Code)
	assert.Nil(t, newUser)
}
func TestUnit_UpdateUser_Deleted(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	u := &User{
		ID:           1,
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	}
	db.DB.Create(u)
	db.DB.Delete(u)

	newUser, err := updateUser(1, &User{
		Email:        "user2@example.com",
		Picture:      "https://example.com/picture2.jpg",
		LastLoginSub: &[]string{"mock-sub2"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(2, 0)}[0],
	})
	assert.Error(t, err)
	assert.Equal(t, apperror.CodeUserNotFound, err.(*apperror.AppError).Code)
	assert.Nil(t, newUser)

	var user User
	db.DB.Unscoped().Where("id = ?", 1).First(&user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "user@example.com", user.Email)
	assert.Equal(t, "https://example.com/picture.jpg", user.Picture)
	assert.Equal(t, "mock-sub", *user.LastLoginSub)
	assert.Equal(t, time.Unix(1, 0), *user.LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), user.CreatedAt)
	assert.Equal(t, time.Unix(1, 0), user.UpdatedAt)
	assert.InDelta(t, time.Now().Unix(), user.DeletedAt.Time.Unix(), 5)
}

func TestUnit_DeleteUser_Normal(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()
	db.DB.Create(&User{
		ID:           1,
		Email:        "user@example.com",
		Picture:      "https://example.com/picture.jpg",
		LastLoginSub: &[]string{"mock-sub"}[0],
		LastLoginAt:  &[]time.Time{time.Unix(1, 0)}[0],
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
	})

	err := deleteUser(1)
	require.NoError(t, err)

	var user User
	db.DB.Unscoped().Where("id = ?", 1).First(&user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "user@example.com", user.Email)
	assert.Equal(t, "https://example.com/picture.jpg", user.Picture)
	assert.Equal(t, "mock-sub", *user.LastLoginSub)
	assert.Equal(t, time.Unix(1, 0), *user.LastLoginAt)
	assert.Equal(t, time.Unix(1, 0), user.CreatedAt)
	assert.Equal(t, time.Unix(1, 0), user.UpdatedAt)
	assert.InDelta(t, time.Now().Unix(), user.DeletedAt.Time.Unix(), 5)
}
func TestUnit_DeleteUser_NotFound(t *testing.T) {
	terminate := test.SetupDB(t, &User{})
	defer terminate()

	err := deleteUser(1)
	require.NoError(t, err)
}
