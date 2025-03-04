package broker

import "context"

type EventType string

const (
	EventTypeProductCreated   EventType = "product.created"
	EventTypeProductUpdated   EventType = "product.updated"
	EventTypeProductDeleted   EventType = "product.deleted"
	EventImageUploaded        EventType = "image.uploaded"
	EventImageProcessed       EventType = "image.processed"
)

type Event interface {
	Type() EventType
}

type ProductEvent struct {
	EventType   EventType `json:"event_type"`
	ProductID   string    `json:"product_id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
}

type ImageEvent struct {
	EventType EventType `json:"event_type"`
	ProductID string    `json:"product_id"`
	ImageData []byte    `json:"image_data"`
}

type MessageBroker interface {
	PublishProduct(ctx context.Context, event *ProductEvent) error
	PublishImage(ctx context.Context, event *ImageEvent) error
	SubscribeToImageProcessed(ctx context.Context, handler func(*ImageEvent) error) error
	Close() error
}
