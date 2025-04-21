package auth

import (
	"context"

	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
)

type AuthService interface {
	Login(ctx context.Context, username, password string) (*pb.LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error)
}