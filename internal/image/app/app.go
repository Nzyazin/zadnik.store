package app

import (
	"context"
    "fmt"
	"log"

    "github.com/Nzyazin/zadnik.store/internal/broker"
    "github.com/Nzyazin/zadnik.store/internal/common"
    "github.com/Nzyazin/zadnik.store/internal/image/storage"
    "github.com/Nzyazin/zadnik.store/internal/image/usecase"
	"github.com/Nzyazin/zadnik.store/internal/image/config"
)

type App struct {
	imageUseCase usecase.ImageUseCase
	messageBroker broker.MessageBroker
	logger common.Logger
}

func NewApp(config *config.Config) (*App, error) {
	logger := common.NewSimpleLogger()

	messageBroker, err := broker.NewRabbitMQBroker(
		broker.RabbitMQConfig{
			URL: config.RabbitMQURL,
		},
	)
	if err != nil {
		log.Fatalf("Failed to initialize message broker: %v", err)
	}

	imageStorage, err := storage.NewFileStorage(
		config.StoragePath,
		config.ImageBaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize file storage: %v", err)
	}

	imageUseCase := usecase.NewImageUseCase(imageStorage, messageBroker, logger)

	return &App{
		imageUseCase: imageUseCase,
		messageBroker: messageBroker,
		logger: logger,
	}, nil
}

func (a *App) handleImageUpload(ctx context.Context, event *broker.ImageEvent) error {
	a.logger.Infof("Received image upload event for product %d", event.ProductID)

	if err := a.imageUseCase.ProcessImage(ctx, event.ImageData, event.ProductID); err != nil {
		a.logger.Errorf("Failed to process image %v", err)
		return err
	}

	a.logger.Infof("Successfully proccessed image for product %d", event.ProductID)
	return nil
}

func (a *App) handleImageDelete(ctx context.Context, event *broker.ProductEvent) error {
	a.logger.Infof("Received product delete event for product %d", event.ProductID)

	if err := a.imageUseCase.DeleteImage(ctx, event.ProductID); err != nil {
		a.logger.Errorf("Failed to delete image for product %d: %v", event.ProductID, err)

		failEvent := &broker.ProductEvent{
			EventType: broker.EventTypeProductDeleted,
			ProductID: event.ProductID,
			Error: err.Error(),
		}
		a.messageBroker.PublishProduct(ctx, failEvent)
		return err
	}

	successEvent := &broker.ProductEvent{
		EventType: broker.EventTypeProductDeleted,
		ProductID: event.ProductID,
	}

	return a.messageBroker.PublishProduct(ctx, successEvent)
}

func (a *App) Run(ctx context.Context) error {
	if err := a.messageBroker.SubscribeToImageUpload(ctx, a.handleImageUpload); err != nil {
		return fmt.Errorf("failed to subscribe to image upload: %w", err)
	}

	if err := a.messageBroker.SubscribeToImageDelete(ctx, a.handleImageDelete); err != nil {
		return fmt.Errorf("failed to subscribe to image delete: %w", err)
	}

	a.logger.Infof("Starting image service")
	<-ctx.Done()
	return nil
}

func (a *App) Shutdown() error {
	return a.messageBroker.Close()
}
