package user

import (
	"log"
	"os"
	"testing"
	"time"

	"vera-identity-service/internal/config"
	"vera-identity-service/internal/db"
	"vera-identity-service/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var d *gorm.DB

func TestMain(m *testing.M) {
	// Setup
	dbURL, closeDB, err := test.SetupPostgresql()
	if err != nil {
		log.Fatal(err)
	}

	d, err = db.NewDatabase(&config.Config{DatabaseURL: dbURL})
	if err != nil {
		log.Fatal(err)
	}
	err = d.AutoMigrate(&User{})
	if err != nil {
		log.Fatal(err)
	}

	// Run
	code := m.Run()

	// Teardown
	closeDB()

	os.Exit(code)
}

func TestRepository_NewRepository_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)

	// Act
	repo := NewRepository(d)

	// Assert
	assert.NotNil(t, repo)
	assert.IsType(t, &repository{}, repo)
}

func TestRepository_GetByID_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{
		ID:           1,
		Name:         test.StringPtr("name"),
		Email:        "test@example.com",
		Picture:      test.StringPtr("https://example.com/picture.png"),
		LastLoginSub: test.StringPtr("mock-login-sub"),
		LastLoginAt:  test.TimePtr(time.Unix(1, 0)),
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
		DeletedAt:    nil,
	}
	err = d.Create(user).Error
	require.NoError(t, err)

	// Act
	result, err := repo.GetByID(user.ID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, user, result)
}
func TestRepository_GetByID_NonExistentUser(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	// Act
	result, err := repo.GetByID(-1)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, result)
}
func TestRepository_GetByID_FilterSoftDeleted(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{
		ID:        1,
		DeletedAt: test.TimePtr(time.Unix(1, 0)),
	}
	err = d.Create(user).Error
	require.NoError(t, err)

	// Act
	result, err := repo.GetByID(user.ID)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRepository_GetByEmail_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{
		ID:           1,
		Name:         test.StringPtr("name"),
		Email:        "test@example.com",
		Picture:      test.StringPtr("https://example.com/picture.png"),
		LastLoginSub: test.StringPtr("mock-login-sub"),
		LastLoginAt:  test.TimePtr(time.Unix(1, 0)),
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
		DeletedAt:    nil,
	}
	err = d.Create(user).Error
	require.NoError(t, err)

	// Act
	result, err := repo.GetByEmail(user.Email)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, user, result)
}
func TestRepository_GetByEmail_EmptyEmail(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	// Act
	result, err := repo.GetByEmail("")

	// Assert
	require.NoError(t, err)
	assert.Nil(t, result)
}
func TestRepository_GetByEmail_NonExistentUser(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	// Act
	result, err := repo.GetByEmail("non-existent-email@example.com")

	// Assert
	require.NoError(t, err)
	assert.Nil(t, result)
}
func TestRepository_GetByEmail_FilterSoftDeleted(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{
		ID:        1,
		DeletedAt: test.TimePtr(time.Unix(1, 0)),
	}
	err = d.Create(user).Error
	require.NoError(t, err)

	// Act
	result, err := repo.GetByEmail(user.Email)

	// Assert
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestRepository_GetAll_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user1 := &User{
		ID:           1,
		Name:         test.StringPtr("name1"),
		Email:        "test1@example.com",
		Picture:      test.StringPtr("https://example.com/picture1.png"),
		LastLoginSub: test.StringPtr("mock-login-sub1"),
		LastLoginAt:  test.TimePtr(time.Unix(1, 0)),
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
		DeletedAt:    nil,
	}
	user2 := &User{
		ID:           2,
		Name:         test.StringPtr("name2"),
		Email:        "test2@example.com",
		Picture:      test.StringPtr("https://example.com/picture2.png"),
		LastLoginSub: test.StringPtr("mock-login-sub2"),
		LastLoginAt:  test.TimePtr(time.Unix(2, 0)),
		CreatedAt:    time.Unix(2, 0),
		UpdatedAt:    time.Unix(2, 0),
		DeletedAt:    nil,
	}
	err = d.Create(user1).Error
	require.NoError(t, err)
	err = d.Create(user2).Error
	require.NoError(t, err)

	// Act
	result, err := repo.GetAll()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []User{*user1, *user2}, result)
}
func TestRepository_GetAll_Empty(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	// Act
	result, err := repo.GetAll()

	// Assert
	require.NoError(t, err)
	assert.Empty(t, result)
}
func TestRepository_GetAll_FilterSoftDeleted(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{
		ID:        1,
		DeletedAt: test.TimePtr(time.Unix(1, 0)),
	}
	err = d.Create(user).Error
	require.NoError(t, err)

	// Act
	result, err := repo.GetAll()

	// Assert
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestRepository_Create_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)
	user := &User{
		Name:         test.StringPtr("name"),
		Email:        "test@example.com",
		Picture:      test.StringPtr("https://example.com/picture.png"),
		LastLoginSub: test.StringPtr("mock-login-sub"),
		LastLoginAt:  test.TimePtr(time.Unix(1, 0)),
	}

	// Act
	err = repo.Create(user)

	// Assert
	require.NoError(t, err)
	expectedUser := &User{
		ID:           1,
		Name:         test.StringPtr("name"),
		Email:        "test@example.com",
		Picture:      test.StringPtr("https://example.com/picture.png"),
		LastLoginSub: test.StringPtr("mock-login-sub"),
		LastLoginAt:  test.TimePtr(time.Unix(1, 0)),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		DeletedAt:    nil,
	}
	assert.Equal(t, expectedUser, user)
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)

	savedUser, err := repo.GetByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, expectedUser, savedUser)
}
func TestRepository_Create_SpecificID(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{
		ID:           10,
		Name:         test.StringPtr("name"),
		Email:        "test@example.com",
		Picture:      test.StringPtr("https://example.com/picture.png"),
		LastLoginSub: test.StringPtr("mock-login-sub"),
		LastLoginAt:  test.TimePtr(time.Unix(1, 0)),
	}

	// Act
	err = repo.Create(user)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, user.ID)

	savedUser, err := repo.GetByID(1)
	require.NoError(t, err)
	assert.NotNil(t, savedUser)
}
func TestRepository_Create_DuplicateID(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{ID: 1}
	err = d.Create(user).Error
	require.NoError(t, err)

	// Act
	err = repo.Create(user)

	// Assert
	require.Error(t, err)

	empty, err := repo.GetByID(user.ID)
	require.NoError(t, err)
	assert.Nil(t, empty)
}

func TestRepository_Update_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{
		ID:           1,
		Name:         test.StringPtr("name"),
		Email:        "test@example.com",
		Picture:      test.StringPtr("https://example.com/picture.png"),
		LastLoginSub: test.StringPtr("mock-login-sub"),
		LastLoginAt:  test.TimePtr(time.Unix(1, 0)),
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
		DeletedAt:    nil,
	}
	err = d.Create(user).Error
	require.NoError(t, err)

	originalUpdatedAt := user.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	user.Name = test.StringPtr("new name")
	user.Email = "new-email@example.com"
	user.Picture = test.StringPtr("https://example.com/new-picture.png")
	user.LastLoginSub = test.StringPtr("new-login-sub")
	user.LastLoginAt = test.TimePtr(time.Unix(2, 0))

	// Act
	err = repo.Update(user)

	// Assert
	require.NoError(t, err)
	expectedUser := &User{
		ID:           1,
		Name:         test.StringPtr("new name"),
		Email:        "new-email@example.com",
		Picture:      test.StringPtr("https://example.com/new-picture.png"),
		LastLoginSub: test.StringPtr("new-login-sub"),
		LastLoginAt:  test.TimePtr(time.Unix(2, 0)),
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    user.UpdatedAt,
		DeletedAt:    nil,
	}
	assert.Equal(t, expectedUser, user)
	assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
	assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

	savedUser, err := repo.GetByID(user.ID)
	require.NoError(t, err)
	assert.Equal(t, expectedUser, savedUser)
}
func TestRepository_Update_NonExistentUser(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{ID: 1}

	// Act
	err = repo.Update(user)

	// Assert
	require.Error(t, err)

	empty, err := repo.GetByID(user.ID)
	require.NoError(t, err)
	assert.Nil(t, empty)
}

func TestRepository_SoftDelete_Success(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	user := &User{
		ID:           1,
		Name:         test.StringPtr("name"),
		Email:        "test@example.com",
		Picture:      test.StringPtr("https://example.com/picture.png"),
		LastLoginSub: test.StringPtr("mock-login-sub"),
		LastLoginAt:  test.TimePtr(time.Unix(1, 0)),
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    time.Unix(1, 0),
		DeletedAt:    nil,
	}
	err = d.Create(user).Error
	require.NoError(t, err)

	// Act
	err = repo.SoftDelete(user.ID)

	// Assert
	require.NoError(t, err)

	deletedUser := &User{}
	err = d.Unscoped().Where("id = ?", user.ID).First(deletedUser).Error
	require.NoError(t, err)
	expectedUser := &User{
		ID:           1,
		Name:         test.StringPtr("name"),
		Email:        "test@example.com",
		Picture:      test.StringPtr("https://example.com/picture.png"),
		LastLoginSub: test.StringPtr("mock-login-sub"),
		LastLoginAt:  test.TimePtr(time.Unix(1, 0)),
		CreatedAt:    time.Unix(1, 0),
		UpdatedAt:    deletedUser.UpdatedAt,
		DeletedAt:    deletedUser.DeletedAt,
	}
	assert.Equal(t, expectedUser, deletedUser)
	assert.WithinDuration(t, time.Now(), deletedUser.UpdatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), *deletedUser.DeletedAt, time.Second)
}
func TestRepository_SoftDelete_NonExistentNode(t *testing.T) {
	// Arrange
	err := test.CleanupTables(d)
	require.NoError(t, err)
	repo := NewRepository(d)

	// Act
	err = repo.SoftDelete(-1)

	// Assert
	require.NoError(t, err)

	count := int64(0)
	err = d.Model(&User{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}
