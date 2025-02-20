package postgres

import (
	"testing"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_GetByID(t *testing.T) {
	db := prepareTestDB(t)
	repo := NewUserRepository(db)

	now := time.Now().UTC().Truncate(time.Second)
	// Создаем тестового пользователя
	user := &domain.User{
		Username:  "testuser",
		PasswordHash:  "hashedpassword",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Вставляем пользователя
	var userID int64
	err := db.QueryRow(
		`INSERT INTO users (username, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	).Scan(&userID)
	require.NoError(t, err)

	t.Run("successful get", func(t *testing.T) {
		foundUser, err := repo.GetByID(userID)
		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.Username, foundUser.Username)
		assert.Equal(t, user.PasswordHash, foundUser.PasswordHash)
		assert.Equal(t, user.CreatedAt.UTC(), foundUser.CreatedAt.UTC())
		assert.Equal(t, user.UpdatedAt.UTC(), foundUser.UpdatedAt.UTC())
	})

	t.Run("user not found", func(t *testing.T) {
		foundUser, err := repo.GetByID(999)
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, foundUser)
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := prepareTestDB(t)
	repo := NewUserRepository(db)

	now := time.Now().UTC().Truncate(time.Second)
	user := &domain.User{
		Username:  "testuser",
		PasswordHash:  "hashedpassword",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Вставляем пользователя
	_, err := db.Exec(
		`INSERT INTO users (username, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`,
		user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt,
	)
	require.NoError(t, err)

	t.Run("successful get", func(t *testing.T) {
		foundUser, err := repo.GetByUsername(user.Username)
		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, user.Username, foundUser.Username)
		assert.Equal(t, user.PasswordHash, foundUser.PasswordHash)
		assert.Equal(t, user.CreatedAt.UTC(), foundUser.CreatedAt.UTC())
		assert.Equal(t, user.UpdatedAt.UTC(), foundUser.UpdatedAt.UTC())
	})

	t.Run("user not found", func(t *testing.T) {
		foundUser, err := repo.GetByUsername("nonexistent")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		assert.Nil(t, foundUser)
	})
}
