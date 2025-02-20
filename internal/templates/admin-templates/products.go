package admin_templates

import (
	"github.com/shopspring/decimal"
)

type Product struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Price decimal.Decimal `json:"price"` 
	Description string `json:"description"`
}
