package main

import (
	"log"
	"net/http"
	"context"

	"github.com/Nzyazin/zadnik.store/internal/product/config"
	"github.com/Nzyazin/zadnik.store/internal/product/repository/postgres"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/product/usecase"
	"github.com/Nzyazin/zadnik.store/internal/product/delivery"
	"github.com/Nzyazin/zadnik.store/internal/broker"

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

	err = messageBroker.SubscribeToImageProcessed(ctx, func(event *broker.ImageEvent) error  {
		logger.Infof("Receiver image processed even for product %s with URL %s", event.ProductID, event.ImageURL)
		
		if err := productRepo.UpdateProductImage(ctx, event.ProductID, event.ImageURL); err != nil {
			logger.Errorf("Failed to update product image: %v", err)
			return err
		}

		logger.Infof("Successfully updated image URL for product %s", event.ProductID)
		return nil
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to image processed events: %v", err)
	}

	router := mux.NewRouter()
	router.Use(productHandler.AuthMiddleware)
	router.HandleFunc("/products", productHandler.GetAll).Methods("GET")
	router.HandleFunc("/products/{id}", productHandler.GetByID).Methods("GET")
	router.HandleFunc("/products/{id}", productHandler.Update).Methods("PATCH")

	logger.Infof("Starting product service on %s", cfg.ProductServiceAddress)
	if err := http.ListenAndServe(cfg.ProductServiceAddress, router); err != nil {
		log.Fatalf("Failed to start product service: %v", err)
	}
}