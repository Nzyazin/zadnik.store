package domain

import (
	"context"
	"database/sql"
	"github.com/shopspring/decimal"
)

type Product struct {
	Name string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Slug string `json:"slug" db:"slug"`
	Price decimal.Decimal `json:"price" db:"price"`
	ImageURL sql.NullString `json:"image_url" db:"image_url"`
	ID int32 `json:"id" db:"id"`
}

type ProductRepository interface {
	GetAll(ctx context.Context) ([]*Product, error)
	GetByID(ctx context.Context, id int32) (*Product, error)
}

