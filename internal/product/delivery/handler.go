package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/product/usecase"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)

type ProductHandler struct {
	productUsecase usecase.ProductUseCase
	logger         common.Logger
	apiKey         string
}

func NewProductHandler(productUsecase usecase.ProductUseCase, logger common.Logger, apiKey string) *ProductHandler {
	return &ProductHandler{
		productUsecase: productUsecase,
		logger:         logger,
		apiKey:         apiKey,
	}
}

func (p *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	p.logger.Infof("Handing GetAll products request")

	products, err := p.productUsecase.GetAll(r.Context())
	if err != nil {
		p.logger.Errorf("Failed to get products: %v", err)
		http.Error(w, "Failed to get products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(products); err != nil {
		p.logger.Errorf("Failed to encode products: %v", err)
		http.Error(w, "Failed to encode products", http.StatusInternalServerError)
		return
	}
}

func (p *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	p.logger.Infof("Handing GetByID product request")

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		p.logger.Errorf("Product ID is empty")
		http.Error(w, "Product ID is empty", http.StatusBadRequest)
		return
	}

	id64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		p.logger.Errorf("Failed to parse product ID: %v", err)
		http.Error(w, "Invalid product ID format", http.StatusBadRequest)
		return
	}

	product, err := p.productUsecase.GetByID(r.Context(), int32(id64))
	if err != nil {
		p.logger.Errorf("Failed to get product: %v", err)
		http.Error(w, "Failed to get product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(product); err != nil {
		p.logger.Errorf("Failed to encode product: %v", err)
		http.Error(w, "Failed to encode product", http.StatusInternalServerError)
		return
	}
}

func (p *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	p.logger.Infof("Handling Update product request")

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		p.logger.Errorf("Product ID is empty")
		http.Error(w, "Product ID is empty", http.StatusBadRequest)
		return
	}

	id64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		p.logger.Errorf("Failed to parse product ID: %v", err)
		http.Error(w, "Invalid product ID format", http.StatusBadRequest)
		return
	}

	currentProduct, err := p.productUsecase.GetByID(r.Context(), int32(id64))
	if err != nil {
		p.logger.Errorf("Failed to get product for update: %v", err)
		http.Error(w, "Failed to get product for update", http.StatusInternalServerError)
		return
	}

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		p.logger.Errorf("Failed to decode update data: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if name, ok := updateData["name"].(string); ok && name != "" {
		currentProduct.Name = name
	}

	if description, ok := updateData["description"].(string); ok && description != "" {
		currentProduct.Description = description
	}

	if slug, ok := updateData["slug"].(string); ok && slug != "" {
		currentProduct.Slug = slug
	}

	if priceStr, ok := updateData["price"].(string); ok && priceStr != "" {
		price, err := decimal.NewFromString(priceStr)
		if err != nil {
			p.logger.Errorf("Failed to parse price: %v", err)
			http.Error(w, "Invalid price format", http.StatusBadRequest)
			return
		}
		currentProduct.Price = price
	} else if priceFloat, ok := updateData["price"].(float64); ok {
		price := decimal.NewFromFloat(priceFloat)
		currentProduct.Price = price
	}

	if imageURL, ok := updateData["image_url"].(string); ok {
		currentProduct.ImageURL.String = imageURL
		currentProduct.ImageURL.Valid = imageURL != ""
	}

	updatedProduct, err := p.productUsecase.Update(r.Context(), currentProduct)
	if err != nil {
		p.logger.Errorf("Failed to updated product: %v", err)
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(updatedProduct); err != nil {
		p.logger.Errorf("Failed to encode updated product: %v", err)
		http.Error(w, "Failed to encode updated product", http.StatusInternalServerError)
		return
	}
}

func (p *ProductHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-KEY")

		if apiKey == "" {
			p.logger.Errorf("API key is empty")
			http.Error(w, "API key is required", http.StatusUnauthorized)
			return
		}

		if apiKey != p.apiKey {
			p.logger.Errorf("Invalid API key provided: %s", apiKey)
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		p.logger.Infof("Request authenticated successfully")
		next.ServeHTTP(w, r)
	})
}
