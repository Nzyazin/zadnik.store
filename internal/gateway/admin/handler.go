package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/broker"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/gateway/auth"
	admin_templates "github.com/Nzyazin/zadnik.store/internal/templates/admin-templates"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type Handler struct {
	authService          auth.AuthService
	templates            *admin_templates.Templates
	productServiceUrl    string
	productServiceAPIKey string
	httpClient           *http.Client
	logger               common.Logger
	messageBroker        broker.MessageBroker
}

func NewHandler(
	authService auth.AuthService,
	templates *admin_templates.Templates,
	productServiceUrl string,
	productServiceAPIKey string,
	messageBroker broker.MessageBroker,
) *Handler {
	return &Handler{
		authService:          authService,
		templates:            templates,
		productServiceUrl:    productServiceUrl,
		productServiceAPIKey: productServiceAPIKey,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		logger:        common.NewSimpleLogger(),
		messageBroker: messageBroker,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	adminGroup := r.Group("/admin")
	{
		adminGroup.GET("/", h.adminIndex)
		// Публичные роуты
		adminGroup.GET("/login", h.loginPage)
		adminGroup.POST("/login", h.login)
		adminGroup.GET("/logout", h.logout)

		// Защищенные роуты
		authorized := adminGroup.Group("/")
		authorized.Use(h.authMiddleware())
		{
			authorized.GET("/products", h.productsIndex)
			authorized.GET("/products/:id/edit", h.productEdit)
			authorized.POST("/products/:id/edit", h.productUpdate)
			authorized.POST("/products/:id/delete", h.productDelete)
		}
	}
}

func (h *Handler) productDelete(c *gin.Context) {
	if !h.checkAuth(c) {
		h.logger.Errorf("Unauthorized attempt to delete product")
		return
	}

	productIDint, err := h.validateProductID(c)
	if err != nil {
		h.logger.Errorf("Product ID validation failed: %v", err)
		h.renderProductsIndex(c, admin_templates.ProductsIndexParams{
			Error: "ID is not valid",
		})
		return
	}

	imageURL := c.PostForm("image_url")

	h.logger.Infof("Starting deletion process for product %d", productIDint)

	productEvent := &broker.ProductEvent{
		EventType: broker.EventTypeProductDeleted,
		ProductID: int32(productIDint),
		ImageURL: imageURL,
	}

	if err := h.messageBroker.PublishProduct(c.Request.Context(), broker.ImageExchange, productEvent); err != nil {
		h.logger.Errorf("Failed to publish product event: %v", err)
		h.renderProductsIndex(c, admin_templates.ProductsIndexParams{
			Error: "Did not can delete product",
		})
		return
	}

	h.logger.Infof("Successfully published delete event for product %d", productIDint)

	// После успешного удаления редиректим на список продуктов
	c.Redirect(http.StatusFound, "/admin/products")
}

func (h *Handler) redirectWithError(c *gin.Context, productID, message string) {
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/products/%s/edit?error=%s",
		productID, url.QueryEscape(message)))
}

func (h *Handler) checkAuth(c *gin.Context) bool {
	_, err := c.Cookie("access_token")
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/login")
		return false
	}
	return true
}

func (h *Handler) validateProductID(c *gin.Context) (int64, error) {
	productID := c.Param("id")
	if productID == "" {
		return 0, fmt.Errorf("product ID is empty")
	}
	return strconv.ParseInt(productID, 10, 64)
}

func (h *Handler) productUpdate(c *gin.Context) {
	if !h.checkAuth(c) {
		return
	}

	productIDInt, err := h.validateProductID(c)
	if err != nil {
		h.logger.Errorf("Product ID validation failed: %v", err)
		c.Redirect(http.StatusFound, "/admin/products")
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")

	originalPrice := c.PostForm("original_price")
	originalName := c.PostForm("original_name")
	originalDescription := c.PostForm("original_description")

	productEvent := &broker.ProductEvent{
		EventType: broker.EventTypeProductUpdated,
		ProductID: int32(productIDInt),
	}
	productIDStr := strconv.FormatInt(productIDInt, 10)

	if priceDecimal, err := h.handlePriceUpdate(priceStr, originalPrice); err != nil {
		h.redirectWithError(c, productIDStr, "Failed to update price")
		return
	} else if priceDecimal != decimal.Zero {
		productEvent.Price = priceDecimal
	}

	if name != originalName {
		productEvent.Name = name
	}

	if description != originalDescription {
		productEvent.Description = description
	}

	if productEvent.Price != decimal.Zero || productEvent.Name != "" || productEvent.Description != "" {
		if err := h.messageBroker.PublishProduct(c.Request.Context(), broker.ProductImageExchange, productEvent); err != nil {
			h.logger.Errorf("Failed to publish product event: %v", err)
			h.redirectWithError(c, productIDStr, "Failed to publish product event")
			return
		}
	}

	if err := h.handleImageUpload(c, productIDInt); err != nil {
		h.redirectWithError(c, strconv.FormatInt(productIDInt, 10), err.Error())
		return
	}

	c.Redirect(http.StatusFound, "/admin/products")
}

func (h *Handler) handleImageUpload(c *gin.Context, productIDInt int64) error {
	file, err := c.FormFile("image")
	if err == http.ErrMissingFile {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found")
	}
	imageData, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer imageData.Close()

	imageBytes, err := io.ReadAll(imageData)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	imageEvent := &broker.ImageEvent{
		EventType: broker.EventTypeImageUploaded,
		ProductID: int32(productIDInt),
		ImageData: imageBytes,
	}

	if err := h.messageBroker.PublishImage(c.Request.Context(), imageEvent); err != nil {
		h.logger.Errorf("Failed to publish image event: %v", err)
		return fmt.Errorf("failed to publish image event: %v", err)
	}

	h.logger.Infof("Successfully published image event for product ID: %d", productIDInt)
	return nil
}

func (h *Handler) handlePriceUpdate(productIDStr, originalPrice string) (decimal.Decimal, error) {
	if productIDStr == originalPrice {
		return decimal.Zero, nil
	}

	priceDecimal, err := decimal.NewFromString(productIDStr)
	if err != nil {
		return decimal.Zero, fmt.Errorf("invalid price format: %w", err)
	}
	if priceDecimal.IsNegative() {
		return decimal.Zero, fmt.Errorf("price cannot be negative")
	}

	if priceDecimal.IsZero() {
		return decimal.Zero, fmt.Errorf("price cannot be zero")
	}

	return priceDecimal, nil
}

func (h *Handler) productEdit(c *gin.Context) {
	_, err := c.Cookie("access_token")
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}

	productID := c.Param("id")
	if productID == "" {
		h.logger.Errorf("Product ID is empty")
		c.Redirect(http.StatusFound, "/admin/products")
		return
	}

	req, err := http.NewRequest(http.MethodGet, h.productServiceUrl+"/products/"+productID, nil)
	if err != nil {
		h.logger.Errorf("Failed to create request: %v", err)
		c.Redirect(http.StatusFound, "/admin/products")
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", h.productServiceAPIKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Errorf("Failed to get product: %v", err)
		c.Redirect(http.StatusFound, "/admin/products")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Errorf("Product service returned non-200 status: %d", resp.Status)
		c.Redirect(http.StatusFound, "/admin/products")
		return
	}

	var product admin_templates.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		h.logger.Errorf("failed to decode product: " + err.Error())
		c.Redirect(http.StatusFound, "/admin/products")
		return
	}

	params := admin_templates.ProductEditParams{
		BaseParams: admin_templates.BaseParams{
			Title: "Редактирование товара - " + product.Name,
		},
		Product: product,
		Error:   c.Query("error"),
	}

	if err := h.templates.RenderProductEdit(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render product template: %v", err)
		c.Redirect(http.StatusFound, "/admin/products")
		return
	}
}

func (h *Handler) adminIndex(c *gin.Context) {
	_, err := c.Cookie("access_token")
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}

	c.Redirect(http.StatusFound, "/admin/products")
}

func (h *Handler) loginPage(c *gin.Context) {
	params := admin_templates.AuthParams{
		Error: c.Query("error"),
	}

	if err := h.templates.RenderAuth(c.Writer, params); err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (h *Handler) login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	resp, err := h.authService.Login(c.Request.Context(), username, password)
	if err != nil {
		params := admin_templates.AuthParams{
			Error: "Неверное имя пользователя или пароль",
		}
		h.templates.RenderAuth(c.Writer, params)
		return
	}

	// Устанавливаем куки
	c.SetCookie(
		"access_token",
		resp.AccessToken,
		int(24*time.Hour.Seconds()),
		"/",
		"",
		false, // secure
		true,  // httpOnly
	)

	c.Redirect(http.StatusFound, "/admin/products")
}

func (h *Handler) logout(c *gin.Context) {
	c.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.Redirect(http.StatusFound, "/admin/login")
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		// Проверяем токен
		resp, err := h.authService.ValidateToken(c.Request.Context(), token)
		if err != nil || !resp.Valid {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		c.Next()
	}
}

func (h *Handler) productsIndex(c *gin.Context) {
	params := admin_templates.ProductsIndexParams{
		BaseParams: admin_templates.BaseParams{
			Title: "Товары",
		},
	}

	req, err := http.NewRequest(http.MethodGet, h.productServiceUrl+"/products", nil)
	if err != nil {
		h.logger.Errorf("Failed to create request: %v", err)
		params.Error = "Не удалось загрузить список товаров"
		h.renderProductsIndex(c, params)
		return
	}
	req.Header.Set("X-API-KEY", h.productServiceAPIKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Errorf("Failed to fetch products: %v", err)
		params.Error = "Не удалось загрузить список товаров"
		h.renderProductsIndex(c, params)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Errorf("Product service returned non-200 status: %d", resp.StatusCode)
		params.Error = "Сервис товаров временно недоступен"
		h.renderProductsIndex(c, params)
		return
	}

	var apiProducts []admin_templates.Product
	if err := json.NewDecoder(resp.Body).Decode(&apiProducts); err != nil {
		h.logger.Errorf("Failed to decode products response: %v", err)
		params.Error = "Ошибка при обработке данных"
		h.renderProductsIndex(c, params)
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
	h.renderProductsIndex(c, params)
}

func (h *Handler) renderProductsIndex(c *gin.Context, params admin_templates.ProductsIndexParams) {
	if err := h.templates.RenderProductsIndex(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render products template: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
	}
}
