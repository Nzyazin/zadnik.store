package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/auth/domain"
	"github.com/Nzyazin/zadnik.store/internal/common"
)

const (
	AccessTokenDuration  = 24 * time.Hour
)

// AuthUseCase определяет все сценарии использования для аутентификации
type AuthUseCase interface {
	Login(ctx context.Context, username, password string) (string, error)
	ValidateAccessToken(ctx context.Context, accessToken string) (int64, error)
	generateAccessToken(userID int64) (string, error)
}

type authUseCase struct {
	userRepo  domain.UserRepository
	logger    common.Logger
	secret    string
}

// NewAuthUseCase создает новый экземпляр AuthUseCase
func NewAuthUseCase(
	userRepo domain.UserRepository,
	logger common.Logger,
	secret string,
) AuthUseCase {
	return &authUseCase{
		userRepo:  userRepo,
		logger:    logger,
		secret:    secret,
	}
}

func (a *authUseCase) Login(ctx context.Context, username, password string) (string, error) {
	user, err := a.userRepo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	if !common.CheckPasswordHash(password, user.PasswordHash) {
		return "", domain.ErrInvalidCredentials
	}

	// Генерируем токены
	return a.generateAccessToken(user.ID)
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
