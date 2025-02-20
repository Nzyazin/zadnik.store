package delivery

import (
	"encoding/json"
	"net/http"
	"github.com/Nzyazin/zadnik.store/internal/product/usecase"
	"github.com/Nzyazin/zadnik.store/internal/common"
)

type ProductHandler struct {
	productUsecase usecase.ProductUseCase
	logger common.Logger
}

func NewProductHandler(productUsecase usecase.ProductUseCase, logger common.Logger) *ProductHandler{
	return &ProductHandler{
		productUsecase: productUsecase,
		logger: logger,
	}
}

func (p *ProductHandler) GetAll(w http.ResponseWriter, r* http.Request) {
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
