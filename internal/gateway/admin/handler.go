package admin

import (
	"net/http"
	"time"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/gateway/auth"
	admin_templates "github.com/Nzyazin/zadnik.store/internal/templates/admin-templates"
)

type Handler struct {
	authService auth.AuthService
	templates   *admin_templates.Templates
	productServiceUrl string
	httpClient  *http.Client
	logger common.Logger
}

func NewHandler(authService auth.AuthService, templates *admin_templates.Templates, productServiceUrl string) *Handler {
	return &Handler{
		authService: authService,
		templates:  templates,
		productServiceUrl: productServiceUrl,
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

type Product struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Price decimal.Decimal `json:"price"`
	Description string `json:"description"`
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

	var apiProducts []Product
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
