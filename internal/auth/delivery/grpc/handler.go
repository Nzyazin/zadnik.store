package grpc

import (
	"context"

	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
	"github.com/Nzyazin/zadnik.store/internal/auth/domain"
	"github.com/Nzyazin/zadnik.store/internal/auth/usecase"
	"github.com/Nzyazin/zadnik.store/internal/common"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authUseCase usecase.AuthUseCase
	logger      common.Logger
}

// NewAuthHandler создает новый экземпляр AuthHandler
func NewAuthHandler(authUseCase usecase.AuthUseCase, logger common.Logger) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		logger:      logger,
	}
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	tokens, err := h.authUseCase.Login(ctx, req.Username, req.Password)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			return nil, domain.ErrInvalidCredentials
		}
		h.logger.Errorf("Login error", "error", err)
		return nil, err
	}

	return &pb.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	tokens, err := h.authUseCase.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Errorf("Refresh token error", "error", err)
		return nil, err
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.authUseCase.Logout(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Errorf("Logout error", "error", err)
		return nil, err
	}

	return &pb.LogoutResponse{}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	userID, err := h.authUseCase.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		h.logger.Errorf("Token validation error", "error", err)
		return nil, err
	}

	return &pb.ValidateTokenResponse{
		UserId: userID,
	}, nil
}
