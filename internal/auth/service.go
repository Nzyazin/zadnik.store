package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const (
	AccessTokenDuration  = 15 * time.Minute
	RefreshTokenDuration = 30 * 24 * time.Hour // 30 days
)

type Service interface {
	Login(ctx context.Context, username, password string) (*TokenPair, error)
	RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
}

type service struct {
	repo   Repository
	logger common.Logger
	secret string
}

func NewService(repo Repository, logger common.Logger, jwtSecret string) Service {
	return &service{
		repo:   repo,
		logger: logger,
		secret: jwtSecret,
	}
}

func (s *service) Login(ctx context.Context, username, password string) (*TokenPair, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		s.logger.Warnf("Failed to get user: %v", err)
		return nil, errors.New("invalid username or password")
	}

	if !common.CheckPasswordHash(password, user.Password) {
		s.logger.Warnf("Invalid password for user: %s", username)
		return nil, errors.New("invalid username or password")
	}

	// Создаем пару токенов
	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		s.logger.Errorf("Failed to generate access token: %v", err)
		return nil, err
	}

	refreshToken := &RefreshToken{
		UserID:    user.ID,
		Token:     uuid.New().String(),
		IsRevoked: false,
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
	}

	if err := s.repo.SaveRefreshToken(ctx, refreshToken); err != nil {
		s.logger.Errorf("Failed to save refresh token: %v", err)
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
	}, nil
}

func (s *service) RefreshTokens(ctx context.Context, refreshToken string) (*TokenPair, error) {
	// Получаем refresh token из БД
	rt, err := s.repo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.Warnf("Invalid refresh token: %v", err)
		return nil, errors.New("invalid refresh token")
	}

	// Отзываем старый refresh token
	if err := s.repo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		s.logger.Errorf("Failed to revoke old refresh token: %v", err)
		return nil, err
	}

	// Генерируем новую пару токенов
	accessToken, err := s.generateAccessToken(rt.UserID)
	if err != nil {
		s.logger.Errorf("Failed to generate access token: %v", err)
		return nil, err
	}

	newRefreshToken := &RefreshToken{
		UserID:    rt.UserID,
		Token:     uuid.New().String(),
		IsRevoked: false,
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
	}

	if err := s.repo.SaveRefreshToken(ctx, newRefreshToken); err != nil {
		s.logger.Errorf("Failed to save new refresh token: %v", err)
		return nil, err
	}

	// Удаляем просроченные токены
	go s.repo.DeleteExpiredTokens(context.Background())

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.Token,
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	if err := s.repo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		s.logger.Warnf("Failed to revoke refresh token: %v", err)
		return err
	}
	return nil
}

func (s *service) generateAccessToken(userID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(AccessTokenDuration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}
