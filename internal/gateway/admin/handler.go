package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Nzyazin/zadnik.store/internal/delivery/http/admin"
)

type Handler struct {
	authService AuthService
}

type AuthService interface {
	// Здесь будут методы для работы с auth service через gRPC
	// Authenticate(ctx context.Context, username, password string) (bool, error)
}

func NewHandler(authService AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	adminGroup := r.Group("/admin")
	{
		// Публичные роуты
		adminGroup.GET("/login", h.loginPage)
		adminGroup.POST("/login", h.login)

		// Защищенные роуты (добавим middleware позже)
		adminGroup.GET("/products", h.productsIndex)
	}
}

func (h *Handler) loginPage(c *gin.Context) {
	params := admin.AuthParams{
		Error: c.Query("error"),
	}
	
	if err := admin.RenderAuth(c.Writer, params); err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func (h *Handler) login(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// TODO: Аутентификация через gRPC
	// authenticated, err := h.authService.Authenticate(c.Request.Context(), username, password)
	
	// Пока просто редирект
	c.Redirect(http.StatusFound, "/admin/products")
}

func (h *Handler) productsIndex(c *gin.Context) {
	// TODO: Получение продуктов через gRPC
	products := []admin.Product{} // пока пустой список

	params := admin.ProductsIndexParams{
		Products: products,
	}
	
	if err := admin.RenderProductsIndex(c.Writer, params); err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
}
