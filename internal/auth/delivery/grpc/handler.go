package grpc

import (
	"context"

	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
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
	h.logger.Infof("Login request received for user: %s", req.Username)
	
	accessToken, err := h.authUseCase.Login(ctx, req.Username, req.Password)
	if err != nil {
		h.logger.Errorf("Login failed: %v", err)
		return nil, err
	}

	return &pb.LoginResponse{
		UserId:      1, // Для единственного админа всегда будет 1
		AccessToken: accessToken,
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	h.logger.Infof("ValidateToken request received")
	
	userID, err := h.authUseCase.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		h.logger.Errorf("Token validation failed: %v", err)
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}
