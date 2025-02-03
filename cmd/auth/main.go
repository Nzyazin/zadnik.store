package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Nzyazin/zadnik.store/internal/auth"
	"github.com/Nzyazin/zadnik.store/internal/auth/config"
	"github.com/Nzyazin/zadnik.store/internal/common"
	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
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

	// Создаем репозиторий
	repo := auth.NewRepository(database)

	// Создаем сервис
	service := auth.NewService(repo, logger, cfg.JWTSecret)

	// Создаем gRPC handler
	handler := auth.NewGRPCHandler(service, logger)

	// Запускаем gRPC сервер
	listener, err := net.Listen("tcp", cfg.AuthServiceAddress)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, handler)

	fmt.Printf("Starting auth service on %s\n", cfg.AuthServiceAddress)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
