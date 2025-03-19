package broker

import (
	"fmt"
	"context"
	"encoding/json"
	"time"

	"github.com/streadway/amqp"
	"github.com/Nzyazin/zadnik.store/internal/common"
)

const (
	ProductExchange = "products"
	ImageExchange = "images"
	ProductImageExchange = "product_images"
)

type RabbitMQBroker struct {
	conn *amqp.Connection
	channel *amqp.Channel
	logger common.Logger
}

type RabbitMQConfig struct {
	URL string
}

func NewRabbitMQBroker(config RabbitMQConfig) (*RabbitMQBroker, error) {
	logger := common.NewSimpleLogger()

	// Используем URL напрямую, так как он уже содержит протокол amqp:// и учетные данные
	rabbitUrl := config.URL

	conn, err := amqp.Dial(rabbitUrl)
	if err != nil {
		logger.Errorf("Failed to connect to RabbitMQ: %v", err)
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Errorf("Failed to create channel: %v", err)
		return nil, fmt.Errorf("failed to oepn a channel: %w", err)
	}
	
	err = declareExchanges(channel)
	if err != nil {
		channel.Close()
		conn.Close()
		logger.Errorf("Failed to declare exchanges: %v", err)
		return nil, fmt.Errorf("failed to declare exchanges: %w", err)
	}

	return &RabbitMQBroker{
		conn: conn,
		channel: channel,
		logger: logger,
	}, nil
}

func declareExchanges(channel *amqp.Channel) error {
	err := channel.ExchangeDeclare(
		ProductExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare product exchange: %w", err)
	}

	err = channel.ExchangeDeclare(
		ImageExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare image exchange: %w", err)
	}

	err = channel.ExchangeDeclare(
		ProductImageExchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare product_image exchange: %w", err)
	}

	return nil
}

func (b *RabbitMQBroker) PublishProduct(ctx context.Context, exchange string, event *ProductEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		b.logger.Errorf("Failed to marshal product event: %v", err)
		return fmt.Errorf("failed to marshal product event: %w", err)
	}

	err = b.channel.Publish(
		exchange,
		string(event.EventType),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body: body,
			DeliveryMode: amqp.Persistent,
			Timestamp: time.Now(),
		})
	
	if err != nil {
		b.logger.Errorf("Failed to publish product event: %v", err)
		return fmt.Errorf("failed to publish product event: %w", err)
	}

	b.logger.Infof("Published product event: %v", event.EventType)
	return nil
}

func (b *RabbitMQBroker) PublishImage(ctx context.Context, event *ImageEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		b.logger.Errorf("Failed to marshal image event: %v", err)
		return fmt.Errorf("failed to marshal image event: %w", err)
	}

	err = b.channel.Publish(
		ImageExchange,
		string(event.EventType),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body: body,
			DeliveryMode: amqp.Persistent,
			Timestamp: time.Now(),
		})
	if err != nil {
		b.logger.Errorf("Failed to publish image event: %v", err)
		return fmt.Errorf("failted to publish image event: %w", err)
	}

	b.logger.Infof("Published image event: %s", event.EventType)
	return nil
}

func (b *RabbitMQBroker) PublishProductImage(ctx context.Context, event *ProductImageEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		b.logger.Errorf("Failed to marshal product image event: %v", err)
		return fmt.Errorf("failed to marshal product image event: %w", err)
	}

	err = b.channel.Publish(
		ImageExchange,
		string(event.EventType),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body: body,
			DeliveryMode: amqp.Persistent,
			Timestamp: time.Now(),
		})
	if err != nil {
		b.logger.Errorf("Failted to publish product image event: %v", err)
		return fmt.Errorf("failed to publish product image event: %w", err)
	}

	b.logger.Infof("Published product image event: %v", event.EventType)
	return nil
}

func (b *RabbitMQBroker) SubscribeToImageProcessed(ctx context.Context, handler func(*ProductImageEvent) error) error {
	queue, err := b.channel.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	
	err = b.channel.QueueBind(
		queue.Name,
		string(EventTypeImageProcessed),
		ImageExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := b.channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				var event ProductImageEvent
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					b.logger.Errorf("Failed to unmarshal image processed event: %v", err)
					continue
				}

				if err := handler(&event); err != nil {
					b.logger.Errorf("Failed to handle image processed event: %v", err)
				}
			}
		}
	}()

	return nil
}

func (b *RabbitMQBroker) SubscribeToProductUpdate(ctx context.Context, handler func(*ProductEvent) error) error {
	queue, err := b.channel.QueueDeclare(
		"",    // имя очереди (пустое для автогенерации)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	
	err = b.channel.QueueBind(
		queue.Name,
		string(EventTypeProductUpdated),
		ProductExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := b.channel.Consume(
		queue.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				var event ProductEvent
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					b.logger.Errorf("Failed to unmarshal product update event: %v", err)
					continue
				}

				if err := handler(&event); err != nil {
					b.logger.Errorf("Failed to handle product update event: %v", err)
				}
			}
		}
	}()

	return nil
}

func (b *RabbitMQBroker) SubscribeToImageUpload(ctx context.Context, handler func(*ImageEvent) error) error {
	b.logger.Infof("Subscribing to image upload events")
	queue, err := b.channel.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}
	
	err = b.channel.QueueBind(
		queue.Name,
		string(EventTypeImageUploaded),
		ImageExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := b.channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to consume messages: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				var event ImageEvent
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					b.logger.Errorf("Failed to unmarshal image upload event: %v", err)
					continue
				}

				if err := handler(&event); err != nil {
					b.logger.Errorf("Failed to handle image upload event: %v", err)
				}
			}
		}
	}()

	return nil
}

func (b *RabbitMQBroker) SubscribeToImageDelete(ctx context.Context, handler func(*ProductEvent) error) error {
	b.logger.Infof("Subscribing to image delete events")
	queue, err := b.channel.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = b.channel.QueueBind(
		queue.Name,
		string(EventTypeProductDeleted),
		ImageExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	msgs, err := b.channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming messages: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					b.logger.Infof("Image delete subscription closed")
					return
				}
				var event ProductEvent
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					b.logger.Errorf("Failed to unmarshal image delete event: %v", err)
					continue
				}

				b.logger.Infof("Received image delete event for product %d", event.ProductID)

				if err := handler(&event); err != nil {
					b.logger.Errorf("Failed to handle product delete event: %v", err)
				}
			}
		}
	}()

	return nil
}

func (b *RabbitMQBroker) Close() error {
	if b.channel != nil {
		b.channel.Close()
	}
	if b.conn != nil {
		b.conn.Close()
	}
	return nil
}
