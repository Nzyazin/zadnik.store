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
	productExchange = "products"
	imageExchange = "images"
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
		productExchange,
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
		imageExchange,
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

	return nil
}

func (b *RabbitMQBroker) PublishProduct(ctx context.Context, event *ProductEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		b.logger.Errorf("Failed to marshal product event: %v", err)
		return fmt.Errorf("failed to marshal product event: %w", err)
	}

	err = b.channel.Publish(
		productExchange,
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
		imageExchange,
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

func (b *RabbitMQBroker) SubscribeToImageProcessed(ctx context.Context, handler func(*ImageEvent) error) error {
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
		string(EventImageProcessed),
		imageExchange,
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

func (b *RabbitMQBroker) Close() error {
	if b.channel != nil {
		b.channel.Close()
	}
	if b.conn != nil {
		b.conn.Close()
	}
	return nil
}
