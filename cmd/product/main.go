package main

import (
	"log"
	"fmt"
	"net/http"

	"github.com/Nzyazin/zadnik.store/internal/product/config"
	"github.com/Nzyazin/zadnik.store/internal/product/repository/postgres"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/product/usecase"
	"github.com/Nzyazin/zadnik.store/internal/product/delivery"

	"github.com/jmoiron/sqlx"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)


func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dbConn, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name))

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	logger := common.NewSimpleLogger()
	productRepo := postgres.NewProductRepository(dbConn)
	productUseCase := usecase.NewProductUseCase(productRepo)
	productHandler := delivery.NewProductHandler(productUseCase, logger)

	router := mux.NewRouter()
	router.HandleFunc("/products", productHandler.GetAll).Methods("GET")

	log.Printf("Starting product service on %s", cfg.ProductServiceAddress)
	if err := http.ListenAndServe(cfg.ProductServiceAddress, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}