package main

import (
	"log"
	"os"
	"path/filepath"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Nzyazin/zadnik.store/internal/broker"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/image/storage"
	"github.com/Nzyazin/zadnik.store/internal/image/usecase"
	"github.com/joho/godotenv"
	"github.com/gin-gonic/gin"
)

func loadEnv() error {
	// Пробуем загрузить .env файл из разных возможных расположений
	envPaths := []string{
		".env-image",
		"internal/image/config/.env-image",
		"../internal/image/config/.env-image",
	}

	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			return nil
		}
	}

	return fmt.Errorf("no .env-image file found in paths: %v", envPaths)
}

func main() {
	if err := loadEnv(); err != nil {
		log.Printf("Warning: %v", err)
		log.Println("Using default environment variables")
	}

	logger := common.NewSimpleLogger()

	// Получаем и валидируем переменные окружения
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "localhost"
	}

	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "/tmp/images"
	}

	// Создаем директорию для хранения если её нет
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Формируем базовый URL для изображений
	imageBaseURL := os.Getenv("IMAGE_BASE_URL")
	if imageBaseURL == "" {
		imageBaseURL = fmt.Sprintf("http://%s:8084/images", domain)
	}

	// Подключаемся к RabbitMQ
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://guest:guest@localhost:5672/"
	}

	messageBroker, err := broker.NewRabbitMQBroker(
		broker.RabbitMQConfig{
			URL: rabbitmqURL,
		},
	)
	if err != nil {
		log.Fatalf("Failed to initialize message broker: %v", err)
	}
	defer messageBroker.Close()

	imageStorage, err := storage.NewFileStorage(
		storagePath,
		imageBaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize file storage: %v", err)
	}

	imageUseCase := usecase.NewImageUseCase(imageStorage, messageBroker, logger)

	// Настраиваем HTTP сервер
	router := gin.Default()

	// Включаем CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Обрабатываем запросы к изображениям
	router.GET("/images/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		if strings.Contains(filename, "..") {
			c.String(http.StatusBadRequest, "Invalid filename")
			return
		}

		imagePath := filepath.Join(storagePath, filename)
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			c.String(http.StatusNotFound, "Image not found")
			return
		}

		c.File(imagePath)
	})

	// Запускаем HTTP сервер в отдельной горутине
	go func() {
		logger.Infof("Starting HTTP server on :8084 (environment: %s, domain: %s)", env, domain)
		if err := router.Run(":8084"); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Подписываемся на события загрузки изображений
	err = messageBroker.SubscribeToImageUpload(ctx, func(event *broker.ImageEvent) error {
		logger.Infof("Received image upload event for product %d", event.ProductID)

		if err := imageUseCase.ProcessImage(ctx, event.ImageData, event.ProductID); err != nil {
			logger.Errorf("Failed to process image %v", err)
			return err
		}

		logger.Infof("Successfully processed image for product %d (storage: %s)", event.ProductID, storagePath)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to image upload: %v", err)
	}

	logger.Infof("Image service is running (storage: %s)", storagePath)
	select {}
}