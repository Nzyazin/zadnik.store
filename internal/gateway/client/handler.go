package client

import (
	"net/http"
	client_templates "github.com/Nzyazin/zadnik.store/internal/templates/client-templates"
	admin_templates "github.com/Nzyazin/zadnik.store/internal/templates/admin-templates"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/gin-gonic/gin"
	"time"
	"encoding/json"
)

type Handler struct {
	templates *client_templates.Templates
	productServiceUrl string
	productServiceAPIKey string
	apiKey string
	logger common.Logger
	httpClient *http.Client
}

func NewHandler(templates *client_templates.Templates, productServiceUrl string, apiKey string) *Handler {
	return &Handler{
		templates: templates,
		productServiceUrl: productServiceUrl,
		apiKey: apiKey,
		logger: common.NewSimpleLogger(),
		httpClient: &http.Client{
			Timeout: time.Second * 9,
		},
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/", h.index)
}

func (h *Handler) index(c *gin.Context) {
	params := client_templates.IndexParams{
		BaseParams: client_templates.BaseParams{
			Title: "Задник из кожкартона саламандер от производителя, доставка по всей России",
            Description: "Задник из кожкартона саламандер от производителя. Доступные цены, 7 видов задника, оптовая продажа с доставкой по России, заказать можно прямо на сайте",
		},
	}
	req, err := http.NewRequest(http.MethodGet, h.productServiceUrl+"/products", nil)
	if err != nil {
		h.logger.Errorf("Failed to create request: %v", err)
		params.Error = "Не удалось загрузить список товаров"
		h.renderIndex(c, params)
		return
	}
	req.Header.Set("X-API-KEY", h.productServiceAPIKey)
	
	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Errorf("Failed to fetch products: %v", err)
		params.Error = "Не удалось загрузить список товаров"
		h.renderIndex(c, params)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Errorf("Product service returned non-200 status: %d", resp.StatusCode)
		params.Error = "Сервис товаров временно недоступен"
		h.renderIndex(c, params)
		return
	}

	var apiProducts []admin_templates.Product
	if err := json.NewDecoder(resp.Body).Decode(&apiProducts); err != nil {
		h.logger.Errorf("Failed to decode products response: %v", err)
		params.Error = "Ошибка при обработке данных"
		h.renderIndex(c, params)
		return
	}

	products := make([]admin_templates.Product, len(apiProducts))
	for i, p := range apiProducts {
		products[i] = admin_templates.Product{
			ID:          p.ID,
			Name:        p.Name,
			Slug:        p.Slug,
			Price:       p.Price,
			Description: p.Description,
			ImageURL:    p.ImageURL,
		}
	}

	params.Products = products
	h.renderIndex(c, params)

}

func (h *Handler) renderIndex(c *gin.Context, params client_templates.IndexParams) {
	if err := h.templates.RenderIndex(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render index template: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Errror")
	}
}