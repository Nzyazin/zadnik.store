package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/auth/domain"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/google/uuid"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 30 * 24 * time.Hour // 30 days
)

// AuthUseCase определяет все сценарии использования для аутентификации
type AuthUseCase interface {
	Login(ctx context.Context, username, password string) (*domain.TokenPair, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateAccessToken(ctx context.Context, accessToken string) (int64, error)
}

type authUseCase struct {
	userRepo  domain.UserRepository
	tokenRepo domain.TokenRepository
	logger    common.Logger
	secret    string
}

// NewAuthUseCase создает новый экземпляр AuthUseCase
func NewAuthUseCase(
	userRepo domain.UserRepository,
	tokenRepo domain.TokenRepository,
	logger common.Logger,
	secret string,
) AuthUseCase {
	return &authUseCase{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		logger:    logger,
		secret:    secret,
	}
}

func (a *authUseCase) Login(ctx context.Context, username, password string) (*domain.TokenPair, error) {
	user, err := a.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if !common.CheckPasswordHash(password, user.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}

	// Генерируем токены
	accessToken, err := a.generateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	err = a.tokenRepo.StoreRefreshToken(&domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
	})

	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authUseCase) RefreshTokens(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	// Получаем информацию о refresh token
	storedToken, err := a.tokenRepo.GetRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Проверяем срок действия
	if time.Now().After(storedToken.ExpiresAt) {
		_ = a.tokenRepo.DeleteRefreshToken(refreshToken)
		return nil, errors.New("refresh token expired")
	}

	// Удаляем старый refresh token
	err = a.tokenRepo.DeleteRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Генерируем новые токены
	accessToken, err := a.generateAccessToken(storedToken.UserID)
	if err != nil {
		return nil, err
	}

	newRefreshToken := uuid.New().String()
	err = a.tokenRepo.StoreRefreshToken(&domain.RefreshToken{
		ID:        uuid.New().String(),
		UserID:    storedToken.UserID,
		Token:     newRefreshToken,
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
	})
	if err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (a *authUseCase) Logout(ctx context.Context, refreshToken string) error {
	return a.tokenRepo.DeleteRefreshToken(refreshToken)
}

func (a *authUseCase) ValidateAccessToken(ctx context.Context, accessToken string) (int64, error) {
	claims, err := common.ValidateJWT(accessToken)
	if err != nil {
		return 0, err
	}

	userID, ok := claims["user_id"].(float64) // JWT преобразует числа в float64
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return int64(userID), nil
}

func (a *authUseCase) generateAccessToken(userID int64) (string, error) {
	claims := map[string]interface{}{
		"user_id": userID,
		"exp":     time.Now().Add(AccessTokenDuration).Unix(),
	}

	return common.GenerateToken(claims, a.secret)
}
