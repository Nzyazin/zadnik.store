package domain

import (
	"context"
)

type Product struct {
	ID int
	Name string
	Slug string
	Description string
	Price string
}

type ProductRepository interface {
	GetList(ctx context.Context) ([]Product, error)
}
