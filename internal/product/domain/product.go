package domain

import (
	"context"
)

type Product struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Slug string `json:"slug"`
	Price string `json:"price"`
	ImageURL string `json:"image_url"`
	ID int32 `json:"id"`
}

type ProductRepository interface {
	GetAll(ctx context.Context) ([]*Product, error)
	GetByID(ctx context.Context, id int32) (*Product, error)
}

