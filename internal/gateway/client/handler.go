package client

import "github.com/Nzyazin/zadnik.store/internal/templates/client_templates"

type Handler struct {
	templates *client_templates.Templates
	productService string
	apiKey string
}

func NewHandler(templates *client_templates.Templates, productService string, apiKey string) *Handler {
	return &Handler{r
		templates: templates,
		productService: productService,
		apiKey: apiKey,
	}
}
		templates: templates,
		productService: productService,
		apiKey: apiKey,
	}
}