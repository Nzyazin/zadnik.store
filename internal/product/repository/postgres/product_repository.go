package postgres

import (
	"github.com/jmoiron/sqlx"
	"context"

	"github.com/Nzyazin/zadnik.store/internal/product/domain"
)

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) GetAll(ctx context.Context) ([]*domain.Product, error) {
	products := []*domain.Product{}
	query := `SELECT * FROM products`
	err := r.db.Select(&products, query)
	return products, err
}