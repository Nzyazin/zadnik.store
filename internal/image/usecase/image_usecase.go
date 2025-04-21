package usecase

import (
	"context"
	"fmt"

	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/image/domain"
)

type ImageUseCase interface {
	CreateImage(ctx context.Context, imageData []byte, filename string, productID int32) (string, error)
	ProcessImage(ctx context.Context, imageData []byte, productID int32) (string, error)
	DeleteImage(ctx context.Context, productID int32) error
}

type imageUseCase struct {
	storage       domain.ImageStorage
	logger        common.Logger
}

func NewImageUseCase(
	storage domain.ImageStorage,
	logger common.Logger,
) ImageUseCase {
	return &imageUseCase{
		storage: storage,
		logger: logger,
	}
}

func (iuc *imageUseCase) CreateImage(ctx context.Context, imageData []byte, filename string, productID int32) (string, error) {
	imageURL, err := iuc.storage.Store(ctx, filename, imageData, productID)
	if err != nil {
		return "", fmt.Errorf("failed to store image: %w", err)
	}

	return imageURL, nil
}

func (iuc *imageUseCase) ProcessImage(ctx context.Context, imageData []byte, productID int32) (string, error) {
	imageURL, err := iuc.storage.Store(ctx, "", imageData, productID)
	if err != nil {
		return "", fmt.Errorf("failed to store image: %w", err)
	}

	return imageURL, nil
}

func (iuc *imageUseCase) DeleteImage(ctx context.Context, productID int32) error {
	imageURL := fmt.Sprintf("%s/%d.jpg", iuc.storage.GetBaseURL(), productID)

	if err := iuc.storage.Delete(ctx, imageURL); err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}
