package main

import (
	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
	"github.com/Nzyazin/zadnik.store/internal/auth"
	"github.com/Nzyazin/zadnik.store/internal/auth/config"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/pkg/db"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	authConfig, err := config.LoadConfig()
	logger := common.NewSimpleLogger()

	if err != nil {
		logger.Errorf("Error connecting to database: %v", err)
	}

	database, err := db.NewFromAuthConfig(authConfig)
	if err != nil {
		logger.Errorf("Error connecting to database: %v", err)
	}
	defer database.Close()

	repo := auth.NewRepository(database)
	service := auth.NewService(repo, logger)
	server := auth.NewGRPCServer(service)

	lis, err := net.Listen("tcp", authConfig.AuthServiceAddress)
	if err != nil {
		logger.Errorf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, server)

	log.Printf("Starting Auth service on %s", authConfig.AuthServiceAddress)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
