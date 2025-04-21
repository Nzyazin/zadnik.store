package auth

import (
	"context"
	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
)
type grpcAuthService struct{
	client pb.AuthServiceClient
}

func NewGRPCAuthService(client pb.AuthServiceClient) AuthService {
	return &grpcAuthService{client: client}
}

func (s *grpcAuthService) Login(ctx context.Context, username, password string) (*pb.LoginResponse, error) {
	return s.client.Login(ctx, &pb.LoginRequest{Username: username, Password: password})
}

func (s *grpcAuthService) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	return s.client.ValidateToken(ctx, &pb.ValidateTokenRequest{AccessToken: token})
}
