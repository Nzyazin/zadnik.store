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
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем логгер
	logger := common.NewSimpleLogger()

	// Подключаемся к базе данных
	database, err := db.NewDatabase(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Инициализируем репозитории
	userRepo := postgres.NewUserRepository(database.DB)
	tokenRepo := postgres.NewTokenRepository(database.DB)

	// Инициализируем use case
	authUseCase := usecase.NewAuthUseCase(userRepo, tokenRepo, logger, cfg.JWTSecret)

	// Инициализируем gRPC handler
	authHandler := authgrpc.NewAuthHandler(authUseCase, logger)

	// Создаем и запускаем gRPC сервер
	listener, err := net.Listen("tcp", cfg.AuthServiceAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, authHandler)

	logger.Infof("Starting gRPC server", "port", cfg.AuthServiceAddress)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
