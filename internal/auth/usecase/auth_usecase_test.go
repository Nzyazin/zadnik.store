package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/auth/domain"
	"github.com/Nzyazin/zadnik.store/internal/auth/mocks"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCase_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	tokenRepo := mocks.NewMockTokenRepository(ctrl)
	logger := common.NewSimpleLogger()
	secret := "test-secret"

	useCase := NewAuthUseCase(userRepo, tokenRepo, logger, secret)

	// Тестовые данные
	username := "testuser"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	user := &domain.User{
		ID:       1,
		Username: username,
		Password: string(hashedPassword),
	}

	t.Run("successful login", func(t *testing.T) {
		// Настраиваем ожидания
		userRepo.EXPECT().
			GetByUsername(username).
			Return(user, nil)

		tokenRepo.EXPECT().
			StoreRefreshToken(gomock.Any()).
			Return(nil)

		// Выполняем тест
		tokens, err := useCase.Login(context.Background(), username, password)

		// Проверяем результаты
		assert.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		// Настраиваем ожидания
		userRepo.EXPECT().
			GetByUsername(username).
			Return(user, nil)

		// Выполняем тест с неверным паролем
		tokens, err := useCase.Login(context.Background(), username, "wrongpassword")

		// Проверяем результаты
		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
	})

	t.Run("user not found", func(t *testing.T) {
		// Настраиваем ожидания
		userRepo.EXPECT().
			GetByUsername(username).
			Return(nil, domain.ErrUserNotFound)

		// Выполняем тест
		tokens, err := useCase.Login(context.Background(), username, password)

		// Проверяем результаты
		assert.Error(t, err)
		assert.Nil(t, tokens)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
	})
}

func TestAuthUseCase_RefreshTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	tokenRepo := mocks.NewMockTokenRepository(ctrl)
	logger := common.NewSimpleLogger()
	secret := "test-secret"

	useCase := NewAuthUseCase(userRepo, tokenRepo, logger, secret)

	refreshToken := "valid-refresh-token"
	user := &domain.User{ID: 1}
	storedToken := &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	t.Run("successful refresh", func(t *testing.T) {
		// Настраиваем ожидания
		tokenRepo.EXPECT().
			GetRefreshToken(refreshToken).
			Return(storedToken, nil)

		tokenRepo.EXPECT().
			DeleteRefreshToken(refreshToken).
			Return(nil)

		tokenRepo.EXPECT().
			StoreRefreshToken(gomock.Any()).
			Return(nil)

		// Выполняем тест
		tokens, err := useCase.RefreshTokens(context.Background(), refreshToken)

		// Проверяем результаты
		assert.NoError(t, err)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		// Настраиваем ожидания
		tokenRepo.EXPECT().
			GetRefreshToken(refreshToken).
			Return(nil, domain.ErrInvalidCredentials)

		// Выполняем тест
		tokens, err := useCase.RefreshTokens(context.Background(), refreshToken)

		// Проверяем результаты
		assert.Error(t, err)
		assert.Nil(t, tokens)
	})
}

func TestAuthUseCase_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := mocks.NewMockUserRepository(ctrl)
	tokenRepo := mocks.NewMockTokenRepository(ctrl)
	logger := common.NewSimpleLogger()
	secret := "test-secret"

	useCase := NewAuthUseCase(userRepo, tokenRepo, logger, secret)

	refreshToken := "valid-refresh-token"

	t.Run("successful logout", func(t *testing.T) {
		// Настраиваем ожидания
		tokenRepo.EXPECT().
			DeleteRefreshToken(refreshToken).
			Return(nil)

		// Выполняем тест
		err := useCase.Logout(context.Background(), refreshToken)

		// Проверяем результаты
		assert.NoError(t, err)
	})

	t.Run("error during logout", func(t *testing.T) {
		// Настраиваем ожидания
		tokenRepo.EXPECT().
			DeleteRefreshToken(refreshToken).
			Return(domain.ErrInvalidCredentials)

		// Выполняем тест
		err := useCase.Logout(context.Background(), refreshToken)

		// Проверяем результаты
		assert.Error(t, err)
	})
}
