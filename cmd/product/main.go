package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/broker"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/product/config"
	"github.com/Nzyazin/zadnik.store/internal/product/delivery"
	"github.com/Nzyazin/zadnik.store/internal/product/repository/postgres"
	"github.com/Nzyazin/zadnik.store/internal/product/server"
	"github.com/Nzyazin/zadnik.store/internal/product/subscriber"
	"github.com/Nzyazin/zadnik.store/internal/product/usecase"

	_ "github.com/lib/pq"
)


func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := postgres.NewPostgresDB(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to initialize db: %v", err)
	}
	defer db.Close()

	logger := common.NewSimpleLogger(&common.LogConfig{FilePath: cfg.LOG_FILE})
	productRepo := postgres.NewProductRepository(db)
	productUseCase := usecase.NewProductUseCase(productRepo)
	productHandler := delivery.NewProductHandler(productUseCase, logger, cfg.APIKey)

	messageBroker, err := broker.NewRabbitMQBroker(broker.RabbitMQConfig{URL: cfg.RabbitMQ.URL, LogFilePath: cfg.LOG_FILE})

	if err != nil {
		log.Fatalf("Failed to initialize message broker: %v", err)
	}
	defer messageBroker.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	subs := subscriber.NewSubscriber(productUseCase, messageBroker, logger)
	if err := subs.Subscribe(ctx); err != nil {
		log.Fatalf("Failed to initialize subscribers: %v", err)
	}

	server := server.NewServer(cfg.ProductServiceAddress, productHandler, logger)

	go func() {
		if err := server.Run(); err != nil {
			logger.Errorf("Server error: %v", err)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 6 * time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}
}