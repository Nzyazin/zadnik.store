package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/Nzyazin/zadnik.store/internal/broker"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/image/storage"
	"github.com/Nzyazin/zadnik.store/internal/image/usecase"
	"github.com/joho/godotenv"
)

func main() {
	projectDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	err = godotenv.Load(filepath.Join("internal", "image", "config", ".env-image"))
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	storagePath := filepath.Join(projectDir, os.Getenv("STORAGE_PATH"))

	logger := common.NewSimpleLogger()

	messageBroker, err := broker.NewRabbitMQBroker(
		broker.RabbitMQConfig{
			URL: os.Getenv("RABBITMQ_URL"),
		},
	)
	if err != nil {
		log.Fatalf("Failed to initialize message broker: %v", err)
	}
	defer messageBroker.Close()

	imageStorage, err := storage.NewFileStorage(
		storagePath,
		os.Getenv("IMAGE_BASE_URL"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize file storage: %v", err)
	}

	imageUseCase := usecase.NewImageUseCase(imageStorage, messageBroker, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = messageBroker.SubscribeToImageUpload(ctx, func(event *broker.ImageEvent) error {
		logger.Infof("Received image upload event for product %d", event.ProductID)

		if err := imageUseCase.ProcessImage(ctx, event.ImageData, event.ProductID); err != nil {
			logger.Errorf("Failed to process image %v", err)
			return err
		}

		logger.Infof("Successfully proccessed image for product %d", event.ProductID)
		return nil
	})

	
	if err != nil {
		log.Fatalf("Failed to subscribe to image upload: %v", err)
	}

	logger.Infof("Starting image service")
	select {}
}
