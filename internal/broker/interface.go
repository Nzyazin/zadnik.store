package broker

import (
	"context"

	"github.com/shopspring/decimal"
)

type EventType string

const (
	EventTypeProductCreated   EventType = "product.created"
	EventTypeProductUpdated   EventType = "product.updated"
	EventTypeProductDeleted   EventType = "product.deleted"
	EventTypeImageUploaded    EventType = "image.uploaded"
	EventTypeImageProcessed   EventType = "image.processed"
)

type Event interface {
	Type() EventType
}

type ProductEvent struct {
	EventType   EventType `json:"event_type"`
	ProductID   int32    `json:"product_id"`
	Name        string    `json:"name"`
	Price       decimal.Decimal   `json:"price"`
	Description string    `json:"description"`
	ImageURL string `json:"image_url"`
	Error string `json:"error,omitempty"`
}

type ImageEvent struct {
	EventType EventType `json:"event_type"`
	ProductID int32    `json:"product_id"`
	ImageData []byte    `json:"image_data"`
}

type ProductImageEvent struct {
	EventType EventType `json:"event_type"`
	ProductID int32    `json:"product_id"`
	ImageURL string    `json:"image_url"`
}

type MessageBroker interface {
	PublishProduct(ctx context.Context, event *ProductEvent) error
	SubscribeToProductUpdate(ctx context.Context, handler func(*ProductEvent) error) error
	PublishImage(ctx context.Context, event *ImageEvent) error
	PublishProductImage(ctx context.Context, event *ProductImageEvent) error
	SubscribeToImageProcessed(ctx context.Context, handler func(*ProductImageEvent) error) error
	SubscribeToImageUpload(ctx context.Context, handler func(*ImageEvent) error) error
	SubscribeToImageDelete(ctx context.Context, handler func(*ProductEvent) error) error
	Close() error
}
