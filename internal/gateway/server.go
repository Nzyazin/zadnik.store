package gateway

import (
	"context"
	"html/template"
	"net/http"
	"time"

	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	router      *gin.Engine
	authClient  pb.AuthServiceClient
	templates   *template.Template
	development bool
}

type ServerConfig struct {
	AuthServiceAddr string
	Development    bool
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	// Подключаемся к сервису auth
	conn, err := grpc.Dial(cfg.AuthServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	// Инициализируем шаблоны
	tmpl, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		return nil, err
	}

	// Создаем сервер
	s := &Server{
		router:      gin.Default(),
		authClient:  pb.NewAuthServiceClient(conn),
		templates:   tmpl,
		development: cfg.Development,
	}

	// Настраиваем маршруты
	s.setupRoutes()

	return s, nil
}

func (s *Server) setupRoutes() {
	// Статические файлы
	s.router.Static("/static", "./web/static")

	// Публичные маршруты
	s.router.GET("/admin/login", s.handleLoginPage)
	s.router.POST("/admin/login", s.handleLogin)
	s.router.POST("/admin/logout", s.handleLogout)

	// Защищенные маршруты
	admin := s.router.Group("/admin")
	admin.Use(s.authMiddleware())
	{
		admin.GET("/", s.handleAdminDashboard)
	}
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) handleLoginPage(c *gin.Context) {
	// Если пользователь уже аутентифицирован, перенаправляем на дашборд
	if token, err := c.Cookie("access_token"); err == nil && token != "" {
		c.Redirect(http.StatusFound, "/admin")
		return
	}

	s.render(c, "login", gin.H{
		"Error":  c.Query("error"),
		"Form":   gin.H{},
		"Errors": gin.H{},
	})
}

func (s *Server) handleLogin(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Вызываем сервис auth
	resp, err := s.authClient.Login(context.Background(), &pb.LoginRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		s.render(c, "login", gin.H{
			"Error": "Invalid username or password",
			"Form": gin.H{
				"username": username,
			},
		})
		return
	}

	// Устанавливаем куки
	c.SetCookie(
		"access_token",
		resp.AccessToken,
		int(time.Hour.Seconds()*24), // 24 часа
		"/",
		"",
		!s.development, // secure cookie in production
		true,          // httpOnly
	)

	c.SetCookie(
		"refresh_token",
		resp.RefreshToken,
		int(time.Hour.Seconds()*24*30), // 30 дней
		"/",
		"",
		!s.development,
		true,
	)

	c.Redirect(http.StatusFound, "/admin")
}

func (s *Server) handleLogout(c *gin.Context) {
	// Получаем refresh token
	refreshToken, err := c.Cookie("refresh_token")
	if err == nil {
		// Вызываем сервис auth для логаута
		_, _ = s.authClient.Logout(context.Background(), &pb.LogoutRequest{
			RefreshToken: refreshToken,
		})
	}

	// Удаляем куки
	c.SetCookie("access_token", "", -1, "/", "", !s.development, true)
	c.SetCookie("refresh_token", "", -1, "/", "", !s.development, true)

	c.Redirect(http.StatusFound, "/admin/login")
}

func (s *Server) handleAdminDashboard(c *gin.Context) {
	// TODO: Implement dashboard page
	c.String(http.StatusOK, "Welcome to admin dashboard!")
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/login")
			c.Abort()
			return
		}

		// Проверяем токен через auth сервис
		resp, err := s.authClient.ValidateToken(context.Background(), &pb.ValidateTokenRequest{
			AccessToken: token,
		})

		if err != nil || !resp.Valid {
			// Пробуем обновить токен
			refreshToken, err := c.Cookie("refresh_token")
			if err != nil {
				c.Redirect(http.StatusFound, "/admin/login")
				c.Abort()
				return
			}

			refreshResp, err := s.authClient.RefreshToken(context.Background(), &pb.RefreshTokenRequest{
				RefreshToken: refreshToken,
			})

			if err != nil {
				c.Redirect(http.StatusFound, "/admin/login")
				c.Abort()
				return
			}

			// Устанавливаем новые куки
			c.SetCookie(
				"access_token",
				refreshResp.AccessToken,
				int(time.Hour.Seconds()*24),
				"/",
				"",
				!s.development,
				true,
			)

			c.SetCookie(
				"refresh_token",
				refreshResp.RefreshToken,
				int(time.Hour.Seconds()*24*30),
				"/",
				"",
				!s.development,
				true,
			)
		}

		// Сохраняем user_id в контексте
		if resp != nil && resp.Valid {
			c.Set("user_id", resp.UserId)
		}

		c.Next()
	}
}

func (s *Server) render(c *gin.Context, name string, data gin.H) {
	if s.development {
		// В режиме разработки перечитываем шаблоны при каждом запросе
		tmpl, err := template.ParseGlob("web/templates/*.html")
		if err == nil {
			s.templates = tmpl
		}
	}

	c.Header("Content-Type", "text/html")
	err := s.templates.ExecuteTemplate(c.Writer, name, data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Template error: "+err.Error())
	}
}
