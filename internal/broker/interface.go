package broker

import (
	"context"

	"github.com/shopspring/decimal"
)

type EventType string

const (
	EventTypeProductCreating   EventType = "product.creating"
	EventTypeProductUpdating   EventType = "product.updating"
	EventTypeProductDeleted   EventType = "product.deleted"
	EventTypeImageUploaded    EventType = "image.uploaded"
	EventTypeImageProcessed   EventType = "image.processed"
	EventTypeImageDeleted EventType = "image.deleted"
	EventTypeImageCreated EventType = "image.created"
	EventTypeProductCreatingCompleted EventType = "product.creating.completed"
	EventTypeProductDeletingCompleted EventType = "product.deleted.completed"
)

type Event interface {
	Type() EventType
}

type ProductEvent struct {
	EventType   EventType `json:"event_type"`
	ProductID   int32    `json:"product_id"`
	ImageData   []byte    `json:"image_data"`
	Name        string    `json:"name"`
	Price       decimal.Decimal   `json:"price"`
	Description string    `json:"description"`
	ImageURL string `json:"image_url"`
	Filename string `json:"filename"`
	Error string `json:"error,omitempty"`
}

func (e *ProductEvent) Type() EventType {
	return e.EventType
}

type ImageEvent struct {
	EventType EventType `json:"event_type"`
	ProductID int32    `json:"product_id"`
	ImageData []byte    `json:"image_data"`
}
func (e *ImageEvent) Type() EventType {
	return e.EventType
}

type ProductImageEvent struct {
	EventType EventType `json:"event_type"`
	ProductID int32    `json:"product_id"`
	ImageURL string    `json:"image_url"`
	Error string `json:"error,omitempty"`
}

func (e *ProductImageEvent) Type() EventType {
	return e.EventType
}

type MessageBroker interface {
	PublishProduct(ctx context.Context, exchange string, event *ProductEvent) error
	PublishImage(ctx context.Context, exchange string, event *ImageEvent) error
	PublishProductImage(ctx context.Context, event *ProductImageEvent) error
	SubscribeToProductUpdate(ctx context.Context, handler func(*ProductEvent) error) error
	SubscribeToImageProcessed(ctx context.Context, handler func(*ProductImageEvent) error) error
	SubscribeToImageUpload(ctx context.Context, exchange string, eventType EventType, handler func(*ImageEvent) error) error
	SubscribeToImageDelete(ctx context.Context, exchange string, eventType EventType, handler func(*ProductEvent) error) error
	SubscribeToImageCreating(ctx context.Context, exchange string, eventType EventType, handler func(*ProductEvent) error) error
	SubscribeToImageCreated(ctx context.Context, exchange string, eventType EventType, handler func(*ProductImageEvent) error) error
	SubscribeToProductDelete(ctx context.Context, exchange string, eventType EventType, handler func(*ProductEvent) error) error
	SubscribeToProductCreated(ctx context.Context, exchange string, eventType EventType, handler func(*ProductEvent) error) error
	SubscribeToProductCreatedCompleted(ctx context.Context, exchange string, eventType EventType, handler func(*ProductEvent) error) error
	Close() error
}
