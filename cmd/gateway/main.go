package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/gateway"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	// Загружаем .env файл
	err := godotenv.Load(filepath.Join("internal", "gateway", "config", ".env-gateway"))
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	logger := common.NewSimpleLogger()

	// Создаем конфигурацию
	cfg := &gateway.ServerConfig{
		AuthServiceAddr: os.Getenv("AUTH_SERVICE_ADDRESS"),
		ProductServiceAddr: os.Getenv("PRODUCT_SERVICE_ADDRESS"),
		ProductServiceAPIKey: os.Getenv("PRODUCT_SERVICE_API_KEY"),
		UserHTTPS: os.Getenv("USE_HTTPS") == "true",
		RabbitMQ: struct {
			URL string
		}{
			URL: os.Getenv("RABBITMQ_URL"),
		},
		Development:    os.Getenv("DEVELOPMENT") == "true",
	}

	// Создаем сервер
	server, err := gateway.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Запускаем сервер
	port := os.Getenv("GATEWAY_PORT")
	logger.Infof("Starting gateway server on :%s\n", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Run(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to run server: %v", err)
		}
	}()

	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
