package gateway

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/Nzyazin/zadnik.store/internal/gateway/admin"
)

type ServerConfig struct {
	AuthServiceAddr string
	Development    bool
}

type Server struct {
	router *gin.Engine
	cfg    *ServerConfig
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	s := &Server{
		router: gin.New(),
		cfg:    cfg,
	}

	// Middleware
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())

	// Static files для админки
	s.router.Static("/admin/assets", "./web/admin/assets")

	// Инициализация сервисов
	authService := NewAuthService(cfg.AuthServiceAddr)

	// Инициализация хендлеров
	adminHandler := admin.NewHandler(authService)
	adminHandler.RegisterRoutes(s.router)

	return s, nil
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

type AuthService struct {
	addr string
}

func NewAuthService(addr string) *AuthService {
	return &AuthService{addr: addr}
}
