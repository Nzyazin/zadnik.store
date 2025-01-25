package auth

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
)

type GRPCServer struct {
	pb.UnimplementedAuthServiceServer
	service Service
}

func NewGRPCServer(service Service) *GRPCServer {
	return &GRPCServer{service: service}
}

func (s *GRPCServer) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	user, token, err := s.service.Authenticate(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid username or password: %v", err)
	}

	return &pb.AuthenticateResponse{
		UserId: user.Username,
		Role:   user.Role,
		Token:  token,
	}, nil
}

func (s *GRPCServer) Authorize(ctx context.Context, req *pb.AuthorizeRequest) (*pb.AuthorizeResponse, error) {
	tokenStr := req.GetToken()
	if tokenStr == "" {
		return nil, status.Errorf(codes.Unauthenticated, "missing token")
	}

	isAuthorized, err := s.service.Authorize(ctx, tokenStr, "admin")
	if err != nil {
		if err.Error() == "forbidden" {
			return nil, status.Errorf(codes.PermissionDenied, "insufficient privileges")
		}
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	return &pb.AuthorizeResponse{
		Authorized: isAuthorized,
	}, nil
}
