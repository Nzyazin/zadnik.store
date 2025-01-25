package auth

import (
	"context"
	"errors"
	"github.com/Nzyazin/zadnik.store/internal/common"
)

type Service interface {
	Authenticate(ctx context.Context, username, password string) (*User, string, error)
	Authorize(ctx context.Context, token, requiredRole string) (bool, error)
}

type service struct {
	repo   Repository
	logger common.Logger
}

func NewService(repo Repository, logger common.Logger) Service {
	return &service{repo: repo, logger: logger}
}

func (s *service) Authenticate(ctx context.Context, username, password string) (*User, string, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			s.logger.Warnf("User not found: %s", username)
			return nil, "", errors.New("invalid username or password")
		}
		s.logger.Errorf("User not found: %s", err)
		return nil, "", err
	}

	if !common.CheckPasswordHash(password, user.Password) {
		s.logger.Warnf("Invalid password for user: %s", username)
		return nil, "", errors.New("invalid username or password")
	}

	token, err := common.GenerateToken(user.ID, user.Role)
	if err != nil {
		s.logger.Errorf("Error generating token for user: %v", err)
		return nil, "", err
	}

	if err := s.repo.SaveToken(ctx, &Token{
		UserID: user.ID,
		Token:  token,
	}); err != nil {
		s.logger.Errorf("Error saving token: %v", err)
		return nil, "", err
	}

	return user, token, nil
}

func (s *service) Authorize(ctx context.Context, token, requiredRole string) (bool, error) {
	user, err := s.repo.GetUserByToken(ctx, token)
	if err != nil {
		if errors.Is(err, common.ErrNotFound) {
			s.logger.Warnf("Token not found or expired")
			return false, errors.New("unauthorized")
		}
		s.logger.Errorf("Error fetching user by token: %v", err)
		return false, err
	}

	if user.Role != requiredRole {
		s.logger.Warnf("Invalid role for user: %s", user.Username)
		return false, errors.New("forbidden")
	}

	return true, nil
}
