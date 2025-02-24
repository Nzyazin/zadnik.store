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
	err := r.db.SelectContext(ctx, &products, query)
	return products, err
}

func (r *productRepository) GetByID(ctx context.Context, id int32) (*domain.Product, error) {
	product := &domain.Product{}
	query := `SELECT * FROM products WHERE id = $1`
	err := r.db.GetContext(ctx, product, query, id)
	return product, err
}