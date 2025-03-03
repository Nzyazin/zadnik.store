package main

import (
	"log"
	"net/http"

	"github.com/Nzyazin/zadnik.store/internal/product/config"
	"github.com/Nzyazin/zadnik.store/internal/product/repository/postgres"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/product/usecase"
	"github.com/Nzyazin/zadnik.store/internal/product/delivery"

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