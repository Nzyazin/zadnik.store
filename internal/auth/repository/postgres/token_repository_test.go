package postgres

import (
	"testing"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/auth/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenRepository_StoreAndGetRefreshToken(t *testing.T) {
	db := prepareTestDB(t)
	repo := NewTokenRepository(db)

	now := time.Now().UTC()
	// Создаем тестового пользователя
	var userID int64
	err := db.QueryRow(
		`INSERT INTO users (username, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		"testuser", "hashedpassword", now, now,
	).Scan(&userID)
	require.NoError(t, err)

	expiresAt := now.Add(24 * time.Hour).Truncate(time.Second)
	token := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     "test-refresh-token",
		ExpiresAt: expiresAt,
	}

	t.Run("store and get token", func(t *testing.T) {
		// Сохраняем токен
		err := repo.StoreRefreshToken(token)
		assert.NoError(t, err)

		// Получаем токен
		foundToken, err := repo.GetRefreshToken(token.Token)
		assert.NoError(t, err)
		assert.NotNil(t, foundToken)
		assert.Equal(t, token.ID, foundToken.ID)
		assert.Equal(t, token.UserID, foundToken.UserID)
		assert.Equal(t, token.Token, foundToken.Token)
		assert.Equal(t, token.ExpiresAt.UTC(), foundToken.ExpiresAt.UTC())
	})

	t.Run("token not found", func(t *testing.T) {
		foundToken, err := repo.GetRefreshToken("nonexistent-token")
		assert.Error(t, err)
		assert.Nil(t, foundToken)
	})
}

func TestTokenRepository_DeleteRefreshToken(t *testing.T) {
	db := prepareTestDB(t)
	repo := NewTokenRepository(db)

	now := time.Now().UTC()
	// Создаем тестового пользователя
	var userID int64
	err := db.QueryRow(
		`INSERT INTO users (username, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		"testuser", "hashedpassword", now, now,
	).Scan(&userID)
	require.NoError(t, err)

	expiresAt := now.Add(24 * time.Hour).Truncate(time.Second)
	token := &domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     "test-refresh-token",
		ExpiresAt: expiresAt,
	}

	t.Run("successful delete", func(t *testing.T) {
		// Сначала сохраняем токен
		err := repo.StoreRefreshToken(token)
		require.NoError(t, err)

		// Удаляем токен
		err = repo.DeleteRefreshToken(token.Token)
		assert.NoError(t, err)

		// Проверяем что токен удален
		foundToken, err := repo.GetRefreshToken(token.Token)
		assert.Error(t, err)
		assert.Nil(t, foundToken)
	})

	t.Run("delete nonexistent token", func(t *testing.T) {
		err := repo.DeleteRefreshToken("nonexistent-token")
		assert.NoError(t, err) // Удаление несуществующего токена не должно возвращать ошибку
	})
}
