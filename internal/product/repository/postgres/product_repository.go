package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/Nzyazin/zadnik.store/internal/product/config"
	"github.com/Nzyazin/zadnik.store/internal/product/domain"
)

type productRepository struct {
	db *sqlx.DB
}

func NewPostgresDB(dbCfg *config.DBConfig) (*sqlx.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Name,
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
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
