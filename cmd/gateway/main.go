package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Nzyazin/zadnik.store/internal/gateway"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
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
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
