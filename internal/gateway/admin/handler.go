package admin

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	admin_templates "github.com/Nzyazin/zadnik.store/internal/templates/admin-templates"
	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
)

type Handler struct {
	authService AuthService
	templates   *admin_templates.Templates
}

type AuthService interface {
	Login(ctx context.Context, username, password string) (*pb.LoginResponse, error)
	ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error)
}

func NewHandler(authService AuthService, templates *admin_templates.Templates) *Handler {
	return &Handler{
		authService: authService,
		templates:  templates,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	adminGroup := r.Group("/admin")
	{
		// Публичные роуты
		adminGroup.GET("/login", h.loginPage)
		adminGroup.POST("/login", h.login)

		// Защищенные роуты
		authorized := adminGroup.Group("/")
		authorized.Use(h.authMiddleware())
		{
			authorized.GET("/products", h.productsIndex)
		}
	}
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
		int(time.Hour.Seconds()*24), // 24 часа
		"/",
		"",
		false, // secure
		true,  // httpOnly
	)

	c.Redirect(http.StatusFound, "/admin/products")
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
	// TODO: Получение продуктов через gRPC
	products := []admin_templates.Product{} // пока пустой список

	params := admin_templates.ProductsIndexParams{
		Products: products,
	}
	
	if err := h.templates.RenderProductsIndex(c.Writer, params); err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
}
