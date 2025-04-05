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

func (r *productRepository) UpdateProductImage(ctx context.Context, productID int32, imageURL string) error {
	query := `UPDATE products SET image_url = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, imageURL, productID)
	return err
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	query := `
		UPDATE products 
		SET name = $1, slug = $2, description = $3, price = $4, image_url = $5
		WHERE id = $6
		RETURNING *
	`
	updatedProduct := &domain.Product{}
	err := r.db.GetContext(
		ctx,
		updatedProduct,
		query,
		product.Name,
		product.Slug,
		product.Description,
		product.Price,
		product.ImageURL,
		product.ID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to updated product: %w", err)
	}

	return updatedProduct, nil
}

func (r *productRepository) BeginDelete(ctx context.Context, productID int32) error {

	var product domain.Product
	err := r.db.GetContext(ctx, &product, `SELECT * FROM products WHERE id = $1 FOR UPDATE`, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if product.Status == domain.ProductStatusDeleting {
		return fmt.Errorf("product %d is already deleted", productID)
	}

	_, err = r.db.ExecContext(ctx, `UPDATE products SET status = $1 WHERE ID = $2`, domain.ProductStatusDeleting, productID)
	if err != nil {
		return fmt.Errorf("failed to begin delete product: %w", err)
	}

	return nil
}

func (r *productRepository) CompleteDelete(ctx context.Context, productID int32) error {
	var product domain.Product
	err := r.db.GetContext(ctx, &product, `SELECT * FROM products WHERE id = $1`, productID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	if product.Status != domain.ProductStatusDeleting {
		return fmt.Errorf("product %d is not in deleting status", productID)
	}

	result, err := r.db.ExecContext(ctx, "DELETE FROM products WHERE id = $1 AND status = $2", productID, domain.ProductStatusDeleting)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("product %d was not deleted", productID)
	}

	return nil
}

func (r *productRepository) RollbackDelete(ctx context.Context, productID int32) error {
	result, err := r.db.ExecContext(ctx, 
		"UPDATE products SET status = $1 WHERE id = $2 AND status = $3",
		domain.ProductStatusActive, productID, domain.ProductStatusDeleting)
	if err != nil {
		return fmt.Errorf("failed to rollback status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("product %d status was not rolled back", productID)
	}
	return nil
}

func (r *productRepository) RollbackCreate(ctx context.Context, productID int32) error {
	result, err := r.db.ExecContext(ctx, 
		"DELETE FROM products WHERE id = $1",
		productID)
	if err != nil {
		return fmt.Errorf("failed to rollback creating product: %d: %w", productID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("product %d was not rolled back", productID)
	}
	return nil
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (name, description, price, status, slug)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	_, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		domain.ProductStatusPending,
		product.Slug,
	)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (r *productRepository) CompleteCreate(ctx context.Context, productID int32) error {
	result, err := r.db.ExecContext(ctx,
		"UPDATE products SET status = $1 WHERE id = $2 AND status = $3",
		domain.ProductStatusActive,
		productID,
		domain.ProductStatusPending)
	if  err != nil {
		return fmt.Errorf("failed to complete creating product: %d: %w", productID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("product %d was not completed", productID)
	}
	return nil
}

func (r *productRepository) BeginCreate(ctx context.Context, product *domain.Product) (*domain.Product, error) {
	query := `
		INSERT INTO products (name, description, price, status, slug)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`

	err := r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Description,
		product.Price,
		domain.ProductStatusPending,
		product.Slug,
	).Scan(&product.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create pending product: %w", err)
	}

	return product, nil
}