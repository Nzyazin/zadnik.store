package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/product/usecase"
	"github.com/gorilla/mux"
)

type ProductHandler struct {
	productUsecase usecase.ProductUseCase
	logger         common.Logger
}

func NewProductHandler(productUsecase usecase.ProductUseCase, logger common.Logger) *ProductHandler {
	return &ProductHandler{
		productUsecase: productUsecase,
		logger:         logger,
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

	id64, err := strconv.ParseInt(id, 10, 32)
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
