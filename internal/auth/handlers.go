package auth

import (
	"context"
	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
	"github.com/Nzyazin/zadnik.store/internal/common"
)

type GRPCHandler struct {
	pb.UnimplementedAuthServiceServer
	service Service
	logger  common.Logger
}

func NewGRPCHandler(service Service, logger common.Logger) pb.AuthServiceServer {
	return &GRPCHandler{
		service: service,
		logger:  logger,
	}
}

func (h *GRPCHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	h.logger.Infof("Login request received for user: %s", req.Username)
	
	tokens, err := h.service.Login(ctx, req.Username, req.Password)
	if err != nil {
		h.logger.Errorf("Login failed: %v", err)
		return nil, err
	}

	return &pb.LoginResponse{
		UserId:       "1", // Для единственного админа всегда будет 1
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *GRPCHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	h.logger.Infof("RefreshToken request received")
	
	tokens, err := h.service.RefreshTokens(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Errorf("Token refresh failed: %v", err)
		return nil, err
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *GRPCHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	h.logger.Infof("Logout request received")
	
	err := h.service.Logout(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Errorf("Logout failed: %v", err)
		return &pb.LogoutResponse{Success: false}, err
	}

	return &pb.LogoutResponse{Success: true}, nil
}

func (h *GRPCHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	h.logger.Infof("ValidateToken request received")
	
	userId, err := h.service.ValidateAccessToken(ctx, req.AccessToken)
	if err != nil {
		h.logger.Errorf("Token validation failed: %v", err)
		return &pb.ValidateTokenResponse{Valid: false}, nil
	}

	return &pb.ValidateTokenResponse{
		Valid:  true,
		UserId: userId,
	}, nil
}
