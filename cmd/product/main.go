package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"fmt"
	"errors"

	"github.com/Nzyazin/zadnik.store/internal/broker"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/product/config"
	"github.com/Nzyazin/zadnik.store/internal/product/delivery"
	"github.com/Nzyazin/zadnik.store/internal/product/repository/postgres"
	"github.com/Nzyazin/zadnik.store/internal/product/usecase"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)


func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := postgres.NewPostgresDB(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to initialize db: %v", err)
	}
	defer db.Close()

	logger := common.NewSimpleLogger()
	productRepo := postgres.NewProductRepository(db)
	productUseCase := usecase.NewProductUseCase(productRepo)
	productHandler := delivery.NewProductHandler(productUseCase, logger, cfg.APIKey)

	messageBroker, err := broker.NewRabbitMQBroker(
		broker.RabbitMQConfig{
			URL: cfg.RabbitMQ.URL,
		},
	)

	if err != nil {
		log.Fatalf("Failed to initialize message broker: %v", err)
	}
	defer messageBroker.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	err = messageBroker.SubscribeToImageProcessed(ctx, func(event *broker.ProductImageEvent) error  {
		logger.Infof("Receiver image processed even for product %d with URL %s", event.ProductID, event.ImageURL)
		
		if err := productUseCase.UpdateProductImage(ctx, event.ProductID, event.ImageURL); err != nil {
			logger.Errorf("Failed to update product image: %v", err)
			return err
		}

		logger.Infof("Successfully updated image URL for product %d", event.ProductID)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to image processed events: %v", err)
	}

	err = messageBroker.SubscribeToProductUpdate(ctx, func(event *broker.ProductEvent) error {
		logger.Infof("Received product update event for product %d", event.ProductID)		
		
		product, err := productUseCase.GetByID(ctx, event.ProductID)
		if err != nil {
			logger.Errorf("Failed to get product: %v", err)
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

		_, err = productUseCase.Update(ctx, product)
		if err != nil {
			logger.Errorf("Failed to update product: %v", err)
			return err
		}

		logger.Infof("Successfully updated product %d", event.ProductID)
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to subscribe to product updated events: %v", err)
	}

	type deleteResult struct {
		productID int64
		err error
	}

	chImageProduct := make(chan deleteResult, 1)
	err = messageBroker.SubscribeToProductDelete(ctx, func(event *broker.ProductEvent) error {
		logger.Infof("Started product deletion for product %d", event.ProductID)
		if event.EventType != broker.EventTypeProductDeleted {
			return nil
		}

		if event.ImageURL == "" {
			if err := productUseCase.BeginDelete(ctx, event.ProductID); err != nil {
				return fmt.Errorf("failed to begin delete product: %d: %w", event.ProductID, err)
			}
			if err := productUseCase.CompleteDelete(ctx, event.ProductID); err != nil {
				return fmt.Errorf("failed to complete delete product: %d: %w", event.ProductID, err)
			}
			logger.Infof("Successfully deleted product %d without image", event.ProductID)
			return nil
		}

		if err := productUseCase.BeginDelete(ctx, event.ProductID); err != nil {
			return fmt.Errorf("failed to begin delete product: %d: %w", event.ProductID, err)
		}

		result := <-chImageProduct
		if result.err != nil {
			if err := productUseCase.RollbackDelete(ctx, event.ProductID); err != nil {
				logger.Errorf("Failed to rollback delete product: %d: %v", event.ProductID, err)
			}
			return fmt.Errorf("failed to delete image for product %d: %w", event.ProductID, result.err)
		}

		if err := productUseCase.CompleteDelete(ctx, event.ProductID); err != nil {
			return fmt.Errorf("failed to complete delete product: %d: %w", event.ProductID, err)
		}
		logger.Infof("Successfully deleted product %d", event.ProductID)
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to subscribe to product deleted events: %v", err)
	}

	err = messageBroker.SubscribeToImageDelete(ctx, broker.ImageExchange, string(broker.EventTypeImageDeleted), func(event *broker.ProductEvent) error {
		logger.Infof("Started subscribe for image deletion for product %d", event.ProductID)
		
		if event.EventType != broker.EventTypeImageDeleted {
			return nil
		}

		var deleteErr error
		if event.Error != "" {
			deleteErr = errors.New(event.Error)
		}
		chImageProduct <- deleteResult{
			productID: int64(event.ProductID),
			err: deleteErr,
		}

		if deleteErr == nil {
			logger.Infof("Successfully deleted image for product %d", event.ProductID)
		} else {
			logger.Errorf("Failed to delete image for product %d: %v", event.ProductID, deleteErr)
		}
		
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to subscribe to image deleted events: %v", err)
	}

	router := mux.NewRouter()
	router.Use(productHandler.AuthMiddleware)
	router.HandleFunc("/products", productHandler.GetAll).Methods("GET")
	router.HandleFunc("/products/{id}", productHandler.GetByID).Methods("GET")
	router.HandleFunc("/products/{id}", productHandler.Update).Methods("PATCH")

	srv := &http.Server {
		Addr: cfg.ProductServiceAddress,
		Handler: router,
	}

	go func() {
		logger.Infof("Starting product service on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("Failed to start product server: %v", err)
		}
	}()

	<-signalChan
	logger.Infof("Received shutdown signal")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 6 * time.Second)
	defer shutdownCancel()

	cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	} else {
		logger.Infof("Server shutdown successfully")
	}

	select {
	case <-shutdownCtx.Done():
		logger.Warnf("Shutdown timeout exceeded, forcing exit")
	default:
		logger.Infof("All operations completed gracefully")
	}
}