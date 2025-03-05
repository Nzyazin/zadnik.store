package usecase

import (
	"context"

	"github.com/Nzyazin/zadnik.store/internal/product/domain"
)

type ProductUseCase interface {
	GetAll(ctx context.Context) ([]*domain.Product, error)
	GetByID(ctx context.Context, id int32) (*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) (*domain.Product, error)
	UpdateProductImage(ctx context.Context, productID int32, imageURL []byte) error
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

func (puc *productUseCase) GetByID(ctx context.Context, id int32) (*domain.Product, error) {
	return puc.repo.GetByID(ctx, id)
}

func (puc *productUseCase) UpdateProductImage(ctx context.Context, productID int32, imageURL string) error {
	return puc.repo.UpdateProductImage(ctx, productID, imageURL)
}

func (puc *productUseCase) Update(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	return puc.repo.Update(ctx, product)
}
