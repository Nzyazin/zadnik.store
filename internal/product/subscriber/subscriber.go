package subscriber

import (
	"context"
	"fmt"
	"errors"

	"github.com/Nzyazin/zadnik.store/internal/broker"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/product/usecase"
)

type Subscriber struct {
	useCase       usecase.ProductUseCase
	messageBroker broker.MessageBroker
	logger        common.Logger
}

type result struct {
	productID int64
	err       error
}

func NewSubscriber(useCase usecase.ProductUseCase, messageBroker broker.MessageBroker, logger common.Logger) *Subscriber {
	return &Subscriber{
		useCase:       useCase,
		messageBroker: messageBroker,
		logger:        logger,
	}
}

func (s *Subscriber) Subscribe(ctx context.Context) error {
	chImageProductDelete:= make(chan result, 1)
	chImageProductCreate:= make(chan result, 1)

	if err := s.subscribeToImageProcessed(ctx); err != nil {
		return err
	}
	if err := s.subscribeToProductUpdate(ctx); err != nil {
		return err
	}
	if err := s.subscribeToProductDelete(ctx, chImageProductDelete); err != nil {
		return err
	}
	if err := s.subscribeToImageDelete(ctx, chImageProductDelete); err != nil {
		return err
	}
	if err := s.subscribeToProductCreated(ctx, chImageProductCreate); err != nil {
		return err
	}
	if err := s.subscribeToImageCreated(ctx, chImageProductCreate); err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) subscribeToImageProcessed(ctx context.Context) error {
	return s.messageBroker.SubscribeToImageProcessed(ctx, func(event *broker.ProductImageEvent) error {
		s.logger.Infof("Received image processed event for product %d with URL %s", event.ProductID, event.ImageURL)

		if err := s.useCase.UpdateProductImage(ctx, event.ProductID, event.ImageURL); err != nil {
			s.logger.Errorf("Failed to update product image: %v", err)
			return err
		}

		s.logger.Infof("Successfully updated image URL for product %d", event.ProductID)
		return nil
	})
}

func (s *Subscriber) subscribeToImageCreated(ctx context.Context, chImageProductCreate chan result) error {
	return s.messageBroker.SubscribeToImageCreated(ctx, broker.ImageExchange, broker.EventTypeImageCreated, func(event *broker.ProductImageEvent) error {
		s.logger.Infof("Received image created event for product %d with URL %s", event.ProductID, event.ImageURL)

		if event.EventType != broker.EventTypeImageCreated {
			return nil
		}

		var deleteErr error
		if event.Error != "" {
			deleteErr = errors.New(event.Error)
		}
		chImageProductCreate <- result{
			productID: int64(event.ProductID),
			err: deleteErr,
		}

		if deleteErr == nil {
			s.logger.Infof("Successfully created image for product %d", event.ProductID)
		} else {
			s.logger.Errorf("Failed to create image for product %d: %v", event.ProductID, deleteErr)
		}
		
		return nil
	})
}
	

func (s *Subscriber) subscribeToProductCreated(ctx context.Context, chImageProductCreate chan result) error {
	return s.messageBroker.SubscribeToProductCreated(ctx, broker.ProductImageCreatingExchange, broker.EventTypeProductCreating, func(event *broker.ProductEvent) error {

		s.logger.Infof("Received data product event")

		if event.ImageData == nil {
			if err := s.useCase.CreateFromEvent(ctx, event); err != nil {
				return fmt.Errorf("failed to begin create product: %d: %w", event.ProductID, err)
			}

			completedEvent := &broker.ProductEvent{
				EventType: broker.EventTypeProductCreatingCompleted,
				ProductID: event.ProductID,
			}
			if err := s.messageBroker.PublishProduct(ctx, broker.ProductImageCreatingCompletedExchange, completedEvent); err != nil {
				s.logger.Errorf("Failed to publish create completed event: %v", err)
			}
			s.logger.Infof("Successfully created product %d without image", event.ProductID)
			return nil
		} else {
			product, err := s.useCase.BeginCreate(ctx, event)
			if err != nil {
				return fmt.Errorf("failed to begin create product: %d: %w", event.ProductID, err)
			}

			result := <-chImageProductCreate
			if result.err != nil {
				if err := s.useCase.RollbackCreate(ctx, product.ID); err != nil {
					s.logger.Errorf("Failed to rollback create product: %d: %v", product.ID, err)
				}
				return fmt.Errorf("failed to create image for product %d: %w", product.ID, result.err)
			}

			if err := s.useCase.CompleteCreate(ctx, product.ID); err != nil {
				return fmt.Errorf("failed to complete create product: %d: %w", product.ID, err)
			}

			completedEvent := &broker.ProductEvent{
				EventType: broker.EventTypeProductCreatingCompleted,
				ProductID: product.ID,
			}
			if err := s.messageBroker.PublishProduct(ctx, broker.ProductImageCreatingCompletedExchange, completedEvent); err != nil {
				s.logger.Errorf("Failed to publish create completed event: %v", err)
			}
			s.logger.Infof("Successfully created product %d", product.ID)
			return nil
		}
	})
}

func (s *Subscriber) subscribeToProductUpdate(ctx context.Context) error {
	return s.messageBroker.SubscribeToProductUpdate(ctx, func(event *broker.ProductEvent) error {
		s.logger.Infof("Received product update event for product %d", event.ProductID)

		product, err := s.useCase.GetByID(ctx, event.ProductID)
		if err != nil {
			s.logger.Errorf("Failed to get product: %v", err)
			return err
		}

		if event.Name != "" {
			product.Name = event.Name
		}
		if !event.Price.IsZero() {
			product.Price = event.Price
		}
		if event.Description != "" {
			product.Description = event.Description
		}

		_, err = s.useCase.Update(ctx, product)
		if err != nil {
			s.logger.Errorf("Failed to update product: %v", err)
			return err
		}

		s.logger.Infof("Successfully updated product %d", event.ProductID)
		return nil
	})
}

func (s *Subscriber) subscribeToProductDelete(ctx context.Context, chImageProduct chan result) error {
	return s.messageBroker.SubscribeToProductDelete(ctx, broker.ProductImageDeletingExchange, broker.EventTypeProductDeleted, func(event *broker.ProductEvent) error {
		if event.EventType != broker.EventTypeProductDeleted {
			return nil
		}
		s.logger.Infof("Started product deletion for product %d", event.ProductID)

		if event.ImageURL == "" {
			if err := s.useCase.BeginDelete(ctx, event.ProductID); err != nil {
				return fmt.Errorf("failed to begin delete product: %d: %w", event.ProductID, err)
			}
			if err := s.useCase.CompleteDelete(ctx, event.ProductID); err != nil {
				return fmt.Errorf("failed to complete delete product: %d: %w", event.ProductID, err)
			}
			s.logger.Infof("Successfully deleted product %d without image", event.ProductID)
			return nil
		}

		if err := s.useCase.BeginDelete(ctx, event.ProductID); err != nil {
			return fmt.Errorf("failed to begin delete product: %d: %w", event.ProductID, err)
		}

		result := <-chImageProduct
		if result.err != nil {
			if err := s.useCase.RollbackDelete(ctx, event.ProductID); err != nil {
				s.logger.Errorf("Failed to rollback delete product: %d: %v", event.ProductID, err)
			}
			return fmt.Errorf("failed to delete image for product %d: %w", event.ProductID, result.err)
		}

		if err := s.useCase.CompleteDelete(ctx, event.ProductID); err != nil {
			return fmt.Errorf("failed to complete delete product: %d: %w", event.ProductID, err)
		}

		completedEvent := &broker.ProductEvent{
			EventType: broker.EventTypeProductDeletingCompleted,
			ProductID: event.ProductID,
		}
		if err := s.messageBroker.PublishProduct(ctx, broker.ProductImageDeletingCompletedExchange, completedEvent); err != nil {
			s.logger.Errorf("Failed to publish delete completed event: %v", err)
		}
		s.logger.Infof("Successfully deleted product %d", event.ProductID)
		return nil
	})
}

func (s *Subscriber) subscribeToImageDelete(ctx context.Context, chImageProduct chan result) error {
	return s.messageBroker.SubscribeToImageDelete(ctx, broker.ImageExchange, broker.EventTypeImageDeleted, func(event *broker.ProductEvent) error {
		s.logger.Infof("Started subscribe for image deletion for product %d", event.ProductID)

		if event.EventType != broker.EventTypeImageDeleted {
			return nil
		}

		var deleteErr error
		if event.Error != "" {
			deleteErr = errors.New(event.Error)
		}
		chImageProduct <- result{
			productID: int64(event.ProductID),
			err: deleteErr,
		}

		if deleteErr == nil {
			s.logger.Infof("Successfully deleted image for product %d", event.ProductID)
		} else {
			s.logger.Errorf("Failed to delete image for product %d: %v", event.ProductID, deleteErr)
		}
		
		return nil
	})
}

