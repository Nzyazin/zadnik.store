package client

import (
	"net/http"
	client_templates "github.com/Nzyazin/zadnik.store/internal/templates/client-templates"
)

type Handler struct {
	templates *client_templates.Templates
	productService string
	apiKey string
}

func NewHandler(templates *client_templates.Templates, productService string, apiKey string) *Handler {
	return &Handler{
		templates: templates,
		productService: productService,
		apiKey: apiKey,
	}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	params := client_templates.IndexParams{
		BaseParams: client_templates.BaseParams{
			Title: "Задник из кожкартона саламандер от производителя, доставка по всей России",
            Description: "Задник из кожкартона саламандер от производителя. Доступные цены, 7 видов задника, оптовая продажа с доставкой по России, заказать можно прямо на сайте",
		},
	}
	products, err := h.getProducts()
	if err != nil {
		http
	}

	}
}