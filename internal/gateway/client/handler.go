package client

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/common"
	client_templates "github.com/Nzyazin/zadnik.store/internal/templates/client-templates"
	"github.com/gin-gonic/gin"
)

type EmailSender interface {
	SendOrder(name, phone string) error
}

type Handler struct {
	templates *client_templates.Templates
	productServiceUrl string
	productServiceAPIKey string
	logger common.Logger
	httpClient *http.Client
	emailSender EmailSender
}

func NewHandler(templates *client_templates.Templates, productServiceUrl string, productServiceAPIKey string, emailSender EmailSender) *Handler {
	return &Handler{
		templates: templates,
		productServiceUrl: productServiceUrl,
		productServiceAPIKey: productServiceAPIKey,
		logger: common.NewSimpleLogger(),
		httpClient: &http.Client{
			Timeout: time.Second * 9,
		},
		emailSender: emailSender,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/", h.indexPage)
	r.GET("/delivery", h.deliveryPage)
	r.GET("/payment", h.paymentPage)
	r.GET("/guarantee", h.guaranteePage)
	r.GET("/policy", h.policyPage)
	r.POST("/send-order", h.sendOrder)
}

func (h *Handler) sendOrder(c *gin.Context) {
	name := c.PostForm("name")
	phone := c.PostForm("phone")

	if phone == "" {
		h.renderError(c, "Номер телефона обязателен")
		return
	}

	if !isValidPhone(phone) {
		h.renderError(c, "Неверный формат номера телефона")
		return
	}

	err := h.emailSender.SendOrder(name, phone)
	if err != nil {
		h.logger.Errorf("Failed to send order: %v", err)
		h.renderError(c, "Не удалось отправить заказ. Пожалуйста, попробуйте позже")
		return
	}
	
	h.renderThank(c, name)
}

func (h *Handler) indexPage(c *gin.Context) {
	params := client_templates.IndexParams{
		BaseParams: client_templates.BaseParams{
			Title: "Задник из кожкартона саламандер от производителя для обуви, доставка по всей России",
            Description: "Задник из кожкартона саламандер от производителя для обуви. Доступные цены, 7 видов задника, оптовая продажа с доставкой по России, заказать можно прямо на сайте",
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

	var apiProducts []client_templates.Product
	if err := json.NewDecoder(resp.Body).Decode(&apiProducts); err != nil {
		h.logger.Errorf("Failed to decode products response: %v", err)
		params.Error = "Ошибка при обработке данных"
		h.renderIndex(c, params)
		return
	}

	products := make([]client_templates.Product, len(apiProducts))
	for i, p := range apiProducts {
		products[i] = client_templates.Product{
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

func (h *Handler) deliveryPage(c *gin.Context) {
	params := client_templates.DeliveryParams{
		BaseParams: client_templates.BaseParams{
			Title: "Доставка задников для обуви из кожартона саламандер",
			Description: "Мы предлагаем быструю и надежную доставку задников для обуви из кожартона саламандер по всей России. Выбираем оптимальный способ доставки с учетом срочности и стоимости. Задники тщательно упакованы для сохранности формы и качества.",
		},
	}
	if err := h.templates.RenderDelivery(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render delivery template: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) paymentPage(c *gin.Context) {
	params := client_templates.PaymentParams{
		BaseParams: client_templates.BaseParams{
			Title: "Оплата задников из кожкартона саламандер для обуви",
			Description: "Способы оплаты задников из кожкартона саламандер от производителя для обуви",
		},
	}
	if err := h.templates.RenderPayment(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render payment template: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) guaranteePage(c *gin.Context) {
	params := client_templates.GuaranteeParams{
		BaseParams: client_templates.BaseParams{
			Title: "Гарантия на задники из кожкартона саламандер для обуви",
			Description: "Гарантийные обязательства и порядок возврата задников для обуви из кожкартона саламандер от производителя для обуви",
		},
	}
	if err := h.templates.RenderGuarantee(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render guarantee template: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) policyPage(c *gin.Context) {
	params := client_templates.PolicyParams{
		BaseParams: client_templates.BaseParams{
			Title: "Политика конфиденциальности",
			Description: "Политика конфиденциальности",
		},
	}
	if err := h.templates.RenderPolicy(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render policy template: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) renderIndex(c *gin.Context, params client_templates.IndexParams) {
	if err := h.templates.RenderIndex(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render index template: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Errror")
	}
}

func (h *Handler) renderError(c *gin.Context, message string) {
	params := client_templates.ErrorParams{
		BaseParams: client_templates.BaseParams{
			Title: "Ошибка",
			Description: "Произошла ошибка",
		},
		Message: message,
	}

	if err := h.templates.RenderError(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render error template: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

func (h *Handler) renderThank(c *gin.Context, name string) {
	params := client_templates.ThankParams{
		BaseParams: client_templates.BaseParams{
			Title: "Благодарим за заказ",
			Description: "Мы свяжемся с вами в ближайшее время.",
		},
		Name: name,
	}

	if err := h.templates.RenderThank(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render thank template: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}

