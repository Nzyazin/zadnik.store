package domain

import (
	"context"
	"database/sql"

	"github.com/shopspring/decimal"
)

type ProductStatus string

const (
	ProductStatusActive ProductStatus = "active"
	ProductStatusDeleting ProductStatus = "deleting"
	ProductStatusDeleted ProductStatus = "deleted"
	ProductStatusPending ProductStatus = "pending"
)

type Product struct {
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Slug        string          `json:"slug" db:"slug"`
	Price       decimal.Decimal `json:"price" db:"price"`
	ImageURL    sql.NullString  `json:"image_url" db:"image_url"`
	ID          int32           `json:"id" db:"id"`
	Status ProductStatus `json:"status" db:"status"`
}

type ProductRepository interface {
	GetAll(ctx context.Context) ([]*Product, error)
	GetByID(ctx context.Context, id int32) (*Product, error)
	UpdateProductImage(ctx context.Context, productID int32, imageURL string) error
	Update(ctx context.Context, product *Product) (*Product, error)
	BeginDelete(ctx context.Context, productID int32) error
	CompleteDelete(ctx context.Context, productID int32) error
	RollbackDelete(ctx context.Context, productID int32) error
	CreatePending(ctx context.Context, product *Product) error
}
