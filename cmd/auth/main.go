package main

import (
	"log"
	"net"

	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
	authgrpc "github.com/Nzyazin/zadnik.store/internal/auth/delivery/grpc"
	"github.com/Nzyazin/zadnik.store/internal/auth/config"
	"github.com/Nzyazin/zadnik.store/internal/auth/repository/postgres"
	"github.com/Nzyazin/zadnik.store/internal/auth/usecase"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/pkg/db"
	"google.golang.org/grpc"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := common.NewSimpleLogger(&common.LogConfig{FilePath: cfg.LOG_FILE})

	database, err := db.NewDatabase(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	userRepo := postgres.NewUserRepository(database.DB)
	authUseCase := usecase.NewAuthUseCase(userRepo, logger, cfg.JWTSecret)
	authHandler := authgrpc.NewAuthHandler(authUseCase, logger)

	listener, err := net.Listen("tcp", cfg.AuthServiceAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, authHandler)

	logger.Infof("Starting gRPC server for authserivce on %s", cfg.AuthServiceAddress)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
