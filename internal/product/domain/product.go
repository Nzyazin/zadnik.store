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
	GetList(ctx context.Context) ([]Product, error)
}
