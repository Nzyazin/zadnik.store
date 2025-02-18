package domain

import (
	"context"
)

type Product struct {
	ID int
	Price string
	Name string
	Slug string
	Description string
}

type ProductRepository interface {
	GetAll(ctx context.Context) ([]*Product, error)
}

