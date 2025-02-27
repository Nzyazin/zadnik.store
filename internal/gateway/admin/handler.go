package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"io"
	"time"
	"mime/multipart"

	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/gateway/auth"
	admin_templates "github.com/Nzyazin/zadnik.store/internal/templates/admin-templates"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService auth.AuthService
	templates   *admin_templates.Templates
	productServiceUrl string
	productServiceAPIKey string
	httpClient  *http.Client
	logger common.Logger
}

func NewHandler(authService auth.AuthService, templates *admin_templates.Templates, productServiceUrl string, productServiceAPIKey string) *Handler {
	return &Handler{
		authService: authService,
		templates:  templates,
		productServiceUrl: productServiceUrl,
		productServiceAPIKey: productServiceAPIKey,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		logger: common.NewSimpleLogger(),
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
		}
	}
}

func (h *Handler) redirectWithError(c *gin.Context, productID, message string) {
	c.Redirect(http.StatusFound, fmt.Sprintf("/admin/products/%s/edit?error=%s",
		productID, url.QueryEscape(message)))
}

func (h *Handler) prepareMultipartForm(c *gin.Context) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for key, values := range c.Request.PostForm {
		if len(values) > 0 {
			if err := writer.WriteField(key, values[0]); err != nil {
				return nil, "", fmt.Errorf("failed to write field %s: %v", key, err)
			}
		}
	}

	file, err := c.FormFile("image"); 
	if err == nil && file != nil {
		if err := h.addFileToForm(writer, file); err != nil {
			return nil, "", err
		}
	}
	
	return body, writer.FormDataContentType(), nil
}

func (h *Handler) addFileToForm(writer *multipart.Writer, file *multipart.FileHeader) error {
	part, err := writer.CreateFormFile("image", file.Filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close()

	if _, err = io.Copy(part, src); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	return nil
}

func (h *Handler) sendProductRequest(productID string, body *bytes.Buffer, contentType string) error {
	req, err := http.NewRequest(http.MethodPatch, h.productServiceUrl + "/products/" + productID, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-API-KEY", h.productServiceAPIKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("product service returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

func (h *Handler) productUpdate(c *gin.Context) {
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

	body, contentType, err := h.prepareMultipartForm(c)
	if err != nil {
		h.logger.Errorf("Failed to prepare form: %v", err)
		h.redirectWithError(c, productID, "Failed to prepare form")
		return
	}

	if err := h.sendProductRequest(productID, body, contentType); err != nil {
		h.logger.Errorf("Failed to update product: %v", err)
		h.redirectWithError(c, productID, "Failed to update product")
		return
	}

	c.Redirect(http.StatusFound, "/admin/products")
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

	resp, err := h.httpClient.Get(h.productServiceUrl + "/products/" + productID)
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
		int(24 * time.Hour.Seconds()),
		"/",
		"",
		false, // secure
		true,  // httpOnly
	)

	c.Redirect(http.StatusFound, "/admin/products")
}

func (h *Handler) logout(c * gin.Context) {
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

	resp, err := h.httpClient.Get(h.productServiceUrl + "/products")
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
			ID: p.ID,
			Name: p.Name,
			Slug: p.Slug,
			Price: p.Price,
			Description: p.Description,
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
