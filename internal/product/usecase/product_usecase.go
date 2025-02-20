package usecase

import (
	"context"
	"github.com/Nzyazin/zadnik.store/internal/product/domain"
)

type ProductUseCase interface {
	GetAll(ctx context.Context) ([]*domain.Product, error)
}

type productUseCase struct {
	repo domain.ProductRepository
}

func NewProductUseCase(repo domain.ProductRepository) ProductUseCase {
	return &productUseCase{repo: repo}
}

func (puc *productUseCase) GetAll(ctx context.Context) ([]*domain.Product, error) {
	return puc.repo.GetAll(ctx)
}
