package domain

import (
	"context"
)

type ImageStorage interface {
	Store(ctx context.Context, imageData []byte, productID int32) (string, error)
	Delete(ctx context.Context, imageURL string) error
	GetBaseURL() string
}
