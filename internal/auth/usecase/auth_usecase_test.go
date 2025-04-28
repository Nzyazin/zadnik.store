package usecase

import (
	"context"
	"testing"

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
	logger := common.NewSimpleLogger()
	secret := "test-secret"

	useCase := NewAuthUseCase(userRepo, logger, secret)

	// Тестовые данные
	username := "testuser"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	
	user := &domain.User{
		ID:       1,
		Username: username,
		PasswordHash: string(hashedPassword),
	}

	t.Run("successful login", func(t *testing.T) {
		// Настраиваем ожидания
		userRepo.EXPECT().
			GetByUsername(username).
			Return(user, nil)

		// Выполняем тест
		tokens, err := useCase.Login(context.Background(), username, password)

		// Проверяем результаты
		assert.NoError(t, err)
		assert.NotEmpty(t, tokens)
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
		assert.Empty(t, tokens)
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
		assert.Empty(t, tokens)
		assert.Equal(t, domain.ErrInvalidCredentials, err)
	})
}
