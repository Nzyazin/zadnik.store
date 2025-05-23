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


const (
    ProductsPath           = "/admin/products"
    ProductCreatePath      = "/admin/products/create"
    ProductEditPathFormat  = "/admin/products/%d/edit"
    ProductDeletePathFormat = "/admin/products/%d/delete"
    
    LoginPath              = "/admin/login"
    LogoutPath             = "/admin/logout"
    
    AdminIndexPath         = "/admin"
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
			Timeout: time.Second * 9,
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
			authorized.GET("/products/create", h.productCreatePage)
			authorized.POST("/products/create", h.productCreate)
			authorized.GET("/products/:id/edit", h.productEditPage)
			authorized.POST("/products/:id/edit", h.productUpdate)
			authorized.POST("/products/:id/delete", h.productDelete)
		}
	}
}

func (h *Handler) redirectWithError(c *gin.Context, productID, message string) {
	if productID == "" {
		c.Redirect(http.StatusFound, ProductCreatePath + "?error=" + url.QueryEscape(message))
		return
	}
	productIDInt, err := strconv.Atoi(productID)
	if err != nil {
		c.Redirect(http.StatusFound, ProductsPath)
		return
	}
	c.Redirect(http.StatusFound, fmt.Sprintf(ProductEditPathFormat + "?error=%s",
		productIDInt, url.QueryEscape(message)))
}

func (h *Handler) checkAuth(c *gin.Context) bool {
	_, err := c.Cookie("access_token")
	if err != nil {
		c.Redirect(http.StatusFound, LoginPath)
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

func (h *Handler) productCreatePage(c *gin.Context) {
	if !h.checkAuth(c) {
		return
	}

	params := admin_templates.ProductFormPageParams{
		BaseParams: admin_templates.BaseParams{
			Title: "Создание товара",
		},
		Action: ProductCreatePath,
		IsEdit: false,
		Product: &admin_templates.Product{},
		ButtonText: "Создать",
		Error: c.Query("error"),
	}

	if err := h.templates.RenderProductFormPage(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render product create page: %v", err)
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (h *Handler) productCreate(c *gin.Context) {
	if !h.checkAuth(c) {
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")

	if name == "" || description == "" || priceStr == "" {
		h.redirectWithError(c, "", "All fields are required")
		return
	}

	productEvent := &broker.ProductEvent{
		EventType:   broker.EventTypeProductCreating,
		Name:        name,
		Description: description,
	}

	imageReader, filename, err := h.handleImage(c)
	if err != nil {
		h.redirectWithError(c, "", "Failed to handle image")
		return
	}
	if imageReader != nil {
		defer imageReader.Close()
		imageBytes, err := io.ReadAll(imageReader)
		if err != nil {
			h.redirectWithError(c, "", "Failed to read image")
			return
		}
		productEvent.Filename = filename
		productEvent.ImageData = imageBytes
	}

	if priceDecimal, err := decimal.NewFromString(priceStr); err != nil {
		h.redirectWithError(c, "", "Invalid price format")
		return
	} else if priceDecimal.IsNegative() || priceDecimal.IsZero() {
		h.redirectWithError(c, "", "Invalid price value")
		return
	} else {
		productEvent.Price = priceDecimal
	}

	done := make(chan error, 1)
	if err := h.messageBroker.SubscribeToProductCreatedCompleted(c.Request.Context(), broker.ProductImageCreatingCompletedExchange, broker.EventTypeProductCreatingCompleted, func(pe *broker.ProductEvent) error {
		h.logger.Infof("Received add completed event for product %d", pe.ProductID)
		done <- nil
		return nil
	}); err != nil {
		h.logger.Errorf("Failed to subscribe to product created completed: %v", err)
		return
	}

	if err := h.messageBroker.PublishProduct(c.Request.Context(), broker.ProductImageCreatingExchange, productEvent); err != nil {
		h.logger.Errorf("Failed to publish product event for creating product: %v", err)
		h.redirectWithError(c, "", "Failed to publish product event for creating product")
		return
	}

	h.logger.Infof("Successfully published create event for product")

	select {
	case <-done:
		c.Redirect(http.StatusFound, ProductsPath)
	case <-time.After(3 * time.Second):
		h.renderProductsIndex(c, admin_templates.ProductsIndexParams{
			Error: "Did not can create product",
		})
	case <-c.Request.Context().Done():
		h.renderProductsIndex(c, admin_templates.ProductsIndexParams{
			Error: "Request cancelled",
		})
	}
}

func (h *Handler) productUpdate(c *gin.Context) {
	if !h.checkAuth(c) {
		return
	}

	productIDInt, err := h.validateProductID(c)
	if err != nil {
		h.logger.Errorf("Product ID validation failed: %v", err)
		c.Redirect(http.StatusFound, ProductsPath)
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")

	originalPrice := c.PostForm("original_price")
	originalName := c.PostForm("original_name")
	originalDescription := c.PostForm("original_description")

	productEvent := &broker.ProductEvent{
		EventType: broker.EventTypeProductUpdating,
		ProductID: int32(productIDInt),
	}
	productIDStr := strconv.FormatInt(productIDInt, 10)

	if priceDecimal, err := h.handlePrice(priceStr, originalPrice); err != nil {
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
		if err := h.messageBroker.PublishProduct(c.Request.Context(), broker.ProductImageUpdatingExchange, productEvent); err != nil {
			h.logger.Errorf("Failed to publish product event: %v", err)
			h.redirectWithError(c, productIDStr, "Failed to publish product event")
			return
		}
	}

	if err := h.handleImageUpload(c, productIDInt); err != nil {
		h.redirectWithError(c, strconv.FormatInt(productIDInt, 10), err.Error())
		return
	}

	c.Redirect(http.StatusFound, ProductsPath)
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

	done := make(chan error, 1)
	cleanup := make(chan struct{})
	defer close(cleanup)
	if err := h.messageBroker.SubscribeToProductDelete(c.Request.Context(), broker.ProductImageDeletingCompletedExchange, broker.EventTypeProductDeletingCompleted, func(pe *broker.ProductEvent) error {
		select {
		case <-cleanup:
			return nil
		default:
			if pe.ProductID == int32(productIDint) {
				done <- nil
				return nil
			}
			return nil
		}
	}); err != nil {
		h.logger.Errorf("Failed to subscribe to image delete: %v", err)
		h.renderProductsIndex(c, admin_templates.ProductsIndexParams{
			Error: "failed to set up product deletion",
		})
		return
	}

	productEvent := &broker.ProductEvent{
		EventType: broker.EventTypeProductDeleted,
		ProductID: int32(productIDint),
		ImageURL:  imageURL,
	}

	if err := h.messageBroker.PublishProduct(c.Request.Context(), broker.ProductImageDeletingExchange, productEvent); err != nil {
		h.logger.Errorf("Failed to publish product event: %v", err)
		h.renderProductsIndex(c, admin_templates.ProductsIndexParams{
			Error: "Failed to initiate product deletion",
		})
		return
	}

	h.logger.Infof("Successfully published delete event for product %d", productIDint)

	select {
	case err :=<-done:
		if err != nil {
			h.logger.Errorf("Error during product deletion: %v", err)
			h.renderProductsIndex(c, admin_templates.ProductsIndexParams{
				Error: "Failed to delete product",
			})
			return
		}
		c.Redirect(http.StatusFound, ProductsPath)
	case <-time.After(9 * time.Second):
		h.renderProductsIndex(c, admin_templates.ProductsIndexParams{
			Error: "Product deletion timeout",
		})
	case <-c.Request.Context().Done():
		h.renderProductsIndex(c, admin_templates.ProductsIndexParams{
			Error: "Request cancelled",
		})
	}

}

func (h *Handler) handleImage(c *gin.Context) (io.ReadCloser, string, error) {
	file, err := c.FormFile("image")
	if err == http.ErrMissingFile {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to get image: %w", err)
	}
	if file == nil {
		return nil, "", fmt.Errorf("file not found")
	}

	imageData, err := file.Open()
	if err != nil {
		return nil, "", fmt.Errorf("failed to open image: %w", err)
	}
	return imageData, file.Filename, nil
}

func (h *Handler) handleImageUpload(c *gin.Context, productIDInt int64) error {
	imageReader, _, err := h.handleImage(c)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}
	if imageReader == nil {
		return nil
	}
	imageEvent := &broker.ImageEvent{
		EventType: broker.EventTypeImageUploaded,
		ProductID: int32(productIDInt),
	}

	defer imageReader.Close()
	imageBytes, err := io.ReadAll(imageReader)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}
	imageEvent.ImageData = imageBytes

	if err := h.messageBroker.PublishImage(c.Request.Context(), broker.ImageExchange, imageEvent); err != nil {
		h.logger.Errorf("Failed to publish image event: %v", err)
		return fmt.Errorf("failed to publish image event: %v", err)
	}

	h.logger.Infof("Successfully published image event for product ID: %d", productIDInt)
	return nil
}

func (h *Handler) handlePrice(productIDStr, originalPrice string) (decimal.Decimal, error) {
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

func (h *Handler) productEditPage(c *gin.Context) {
	_, err := c.Cookie("access_token")
	if err != nil {
		c.Redirect(http.StatusFound, LoginPath)
		return
	}

	productID := c.Param("id")
	if productID == "" {
		h.logger.Errorf("Product ID is empty")
		c.Redirect(http.StatusFound, ProductsPath)
		return
	}

	req, err := http.NewRequest(http.MethodGet, h.productServiceUrl+"/products/"+productID, nil)
	if err != nil {
		h.logger.Errorf("Failed to create request: %v", err)
		c.Redirect(http.StatusFound, ProductsPath)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", h.productServiceAPIKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Errorf("Failed to get product: %v", err)
		c.Redirect(http.StatusFound, ProductsPath)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Errorf("Product service returned non-200 status: %d", resp.Status)
		c.Redirect(http.StatusFound, ProductsPath)
		return
	}

	var product admin_templates.Product
	if err := json.NewDecoder(resp.Body).Decode(&product); err != nil {
		h.logger.Errorf("failed to decode product: " + err.Error())
		c.Redirect(http.StatusFound, ProductsPath)
		return
	}

	params := admin_templates.ProductFormPageParams{
		BaseParams: admin_templates.BaseParams{
			Title: "Редактирование товара - " + product.Name,
		},
		Action: fmt.Sprintf(ProductEditPathFormat, product.ID),
		IsEdit: true,
		Product: &product,
		ButtonText: "Сохранить",
		Error:   c.Query("error"),
	}

	if err := h.templates.RenderProductFormPage(c.Writer, params); err != nil {
		h.logger.Errorf("Failed to render product template: %v", err)
		c.Redirect(http.StatusFound, ProductsPath)
		return
	}
}

func (h *Handler) adminIndex(c *gin.Context) {
	_, err := c.Cookie("access_token")
	if err != nil {
		c.Redirect(http.StatusFound, LoginPath)
		return
	}

	c.Redirect(http.StatusFound, ProductsPath)
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

	c.Redirect(http.StatusFound, ProductsPath)
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

	c.Redirect(http.StatusFound, LoginPath)
}

func (h *Handler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.Redirect(http.StatusFound, LoginPath)
			c.Abort()
			return
		}

		// Проверяем токен
		resp, err := h.authService.ValidateToken(c.Request.Context(), token)
		if err != nil || !resp.Valid {
			c.Redirect(http.StatusFound, LoginPath)
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
