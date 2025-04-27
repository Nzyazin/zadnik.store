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
	"strconv"

	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/gateway"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/Nzyazin/zadnik.store/internal/broker"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	// Загружаем .env файл
	err := godotenv.Load(filepath.Join("internal", "gateway", "config", ".env-gateway"))
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	portStrSMTP := os.Getenv("EMAIL_SMTP_PORT")
	portSMTP, err := strconv.Atoi(portStrSMTP)
	if err != nil {
		portSMTP = 587 // Значение по умолчанию, если преобразование не удалось
	}

	// Создаем конфигурацию
	cfg := &gateway.ServerConfig{
		AuthServiceAddr: os.Getenv("AUTH_SERVICE_ADDRESS"),
		ProductServiceAddr: os.Getenv("PRODUCT_SERVICE_ADDRESS"),
		ProductServiceAPIKey: os.Getenv("PRODUCT_SERVICE_API_KEY"),
		UseHTTPS: os.Getenv("USE_HTTPS") == "true",
		RabbitMQ: broker.RabbitMQConfig{
			URL: os.Getenv("RABBITMQ_URL"),
			LogFilePath: os.Getenv("LOG_FILE"),
		},
		Development:    os.Getenv("DEVELOPMENT") == "true",
		SMTPConfig: gateway.SMTPConfig{
			Host: os.Getenv("EMAIL_SMTP_HOST"),
			Port: portSMTP,
			From: os.Getenv("EMAIL_FROM"),
			Password: os.Getenv("EMAIL_PASSWORD"),
		},
		CertFile: os.Getenv("CERT_FILE"),
		KeyFile: os.Getenv("KEY_FILE"),
		LOG_FILE: os.Getenv("LOG_FILE"),
	}

	logger := common.NewSimpleLogger(&common.LogConfig{FilePath: cfg.LOG_FILE})

	// Создаем сервер
	server, err := gateway.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Запускаем сервер
	port := os.Getenv("GATEWAY_PORT")
	gatewayHost := os.Getenv("GATEWAY_HOST")
	if gatewayHost == "" {
		gatewayHost = "localhost"
	}

	useHTTPS := cfg.UseHTTPS
	httpsPort := os.Getenv("HTTPS_PORT")
	if httpsPort == "" {
		if cfg.Development {
			httpsPort = "8443"
		} else {
			httpsPort = "443"
		}
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		if cfg.Development {
			httpPort = "8082"
		} else {
			httpPort = "80"
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if useHTTPS && cfg.CertFile != "" && cfg.KeyFile != "" {
		srv := gateway.RunHTTPRedirect(":" + httpPort, gatewayHost)
		defer srv.Shutdown(context.Background())

		logger.Infof("Starting HTTPS gateway server on %s:%s\n", gatewayHost, httpsPort)
		logger.Infof("HTTP to HTTPS redirect active on port %s\n", httpPort)
		go func() {
			if err := server.RunWithTLS(gatewayHost + ":" + httpsPort, cfg.CertFile, cfg.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to run server: %v", err)
			}
		}()
	} else {
		logger.Infof("Starting HTTP gateway server on %s:%s\n", gatewayHost, port)
		go func() {
			if err := server.Run(gatewayHost + ":" + port); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Failed to run server: %v", err)
			}
		}()
	}

	<-sigChan

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
