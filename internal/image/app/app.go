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
	logger := common.NewSimpleLogger(&common.LogConfig{FilePath: config.LOG_FILE})

	messageBroker, err := broker.NewRabbitMQBroker(
		broker.RabbitMQConfig{
			URL: config.RabbitMQURL,
			LogFilePath: config.LOG_FILE,
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

	imageUseCase := usecase.NewImageUseCase(imageStorage, logger)

	return &App{
		imageUseCase: imageUseCase,
		messageBroker: messageBroker,
		logger: logger,
	}, nil
}

func (a *App) handleImageUpload(event *broker.ImageEvent) error {
	a.logger.Infof("Received image upload event for product %d", event.ProductID)

	ctx := context.Background()

	imageUrl, err := a.imageUseCase.ProcessImage(ctx, event.ImageData, event.ProductID); 
	if err != nil {
		a.logger.Errorf("Failed to process image %v", err)
		return err
	}

	eventFinished := &broker.ProductImageEvent{
		EventType: broker.EventTypeImageProcessed,
		ProductID: event.ProductID,
		ImageURL: imageUrl,
	}

	if err := a.messageBroker.PublishProductImage(ctx, eventFinished); err != nil {
		if delErr := a.imageUseCase.DeleteImage(ctx, event.ProductID); delErr != nil {
			a.logger.Errorf("Failed to delete image: %v", delErr)
		}
		return fmt.Errorf("failed to publish image processed event: %w", err)
	}

	a.logger.Infof("Successfully proccessed image for product %d", event.ProductID)
	return nil
}

func (a *App) handleImageDelete(event *broker.ProductEvent) error {
	a.logger.Infof("Received product delete event for product %d", event.ProductID)

	ctx := context.Background()

	if err := a.imageUseCase.DeleteImage(ctx, event.ProductID); err != nil {
		a.logger.Errorf("Failed to delete image for product %d: %v", event.ProductID, err)

		failEvent := &broker.ProductEvent{
			EventType: broker.EventTypeImageDeleted,
			ProductID: event.ProductID,
			Error: err.Error(),
		}
		a.messageBroker.PublishProduct(ctx, broker.ImageExchange, failEvent)
		return err
	}

	successEvent := &broker.ProductEvent{
		EventType: broker.EventTypeImageDeleted,
		ProductID: event.ProductID,
		Error: "",
	}

	return a.messageBroker.PublishProduct(ctx, broker.ImageExchange, successEvent)
}

func (a *App) handleImageCreating(event *broker.ProductEvent) error {
	a.logger.Infof("Received image creating event for product %d", event.ProductID)

	ctx := context.Background()
	eventFinished := &broker.ProductImageEvent{
		EventType: broker.EventTypeImageCreated,
		ProductID: event.ProductID,
	}

	imageUrl, err := a.imageUseCase.CreateImage(ctx, event.ImageData, event.Filename, event.ProductID); 
	if err != nil {
		a.logger.Errorf("Failed to process image %v", err)
		eventFinished.Error = err.Error()
		if err := a.messageBroker.PublishProductImage(ctx, eventFinished); err != nil {
			if delErr := a.imageUseCase.DeleteImage(ctx, event.ProductID); delErr != nil {
				a.logger.Errorf("Failed to delete image after error publishProductImage: %v", delErr)
			}
			return fmt.Errorf("failed to publish EventTypeImageCreated: %w", err)
		}
		return err
	}

	eventFinished.ImageURL = imageUrl

	if err := a.messageBroker.PublishProductImage(ctx, eventFinished); err != nil {
		if delErr := a.imageUseCase.DeleteImage(ctx, event.ProductID); delErr != nil {
			a.logger.Errorf("Failed to delete image after error publishProductImage: %v", delErr)
		}
		return fmt.Errorf("failed to publish EventTypeImageCreated: %w", err)
	}

	a.logger.Infof("Successfully proccessed image for product %d", event.ProductID)

	return nil
}

func (a *App) Run(ctx context.Context) error {
	if err := a.messageBroker.SubscribeToImageUpload(ctx, broker.ImageExchange, broker.EventTypeImageUploaded, a.handleImageUpload); err != nil {
		return fmt.Errorf("failed to subscribe to image upload: %w", err)
	}

	if err := a.messageBroker.SubscribeToImageDelete(ctx, broker.ProductImageDeletingExchange,  broker.EventTypeProductDeleted, a.handleImageDelete); err != nil {
		return fmt.Errorf("failed to subscribe to image delete: %w", err)
	}

	if err := a.messageBroker.SubscribeToImageCreating(ctx, broker.ProductImageCreatingExchange, broker.EventTypeProductCreating, a.handleImageCreating); err != nil {
		return fmt.Errorf("failed to subscribe to image creating: %w", err)
	}

	a.logger.Infof("Starting image service")
	<-ctx.Done()
	return nil
}

func (a *App) Shutdown() error {
	return a.messageBroker.Close()
}
