package usecase

import (
	"context"
	"fmt"

	"github.com/Nzyazin/zadnik.store/internal/broker"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/image/domain"
)

type ImageUseCase interface {
	ProcessImage(ctx context.Context, imageData []byte, productID int32) error
}

type imageUseCase struct {
	storage       domain.ImageStorage
	messageBroker broker.MessageBroker
	logger        common.Logger
}

func NewImageUseCase(
	storage domain.ImageStorage,
	messageBroker broker.MessageBroker,
	logger common.Logger,
) ImageUseCase {
	return &imageUseCase{
		storage: storage,
		messageBroker: messageBroker,
		logger: logger,
	}
}

func (iuc *imageUseCase) ProcessImage(ctx context.Context, imageData []byte, productID int32) error {
	imageURL, err := iuc.storage.Store(ctx, imageData, productID)
	if err != nil {
		return fmt.Errorf("failed to store image: %w", err)
	}

	event := &broker.ProductImageEvent{
		EventType: broker.EventImageUploaded,
		ProductID: productID,
		ImageURL: imageURL,
	}

	if err := iuc.messageBroker.PublishProductImage(ctx, event); err != nil {
		if delErr := iuc.storage.Delete(ctx, imageURL); delErr != nil {
			iuc.logger.Errorf("Failed to delete image: %v", delErr)
		}
		return fmt.Errorf("failed to publish image processed event: %w", err)
	}
	
	return nil
}
