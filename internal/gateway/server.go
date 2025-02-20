package gateway

import (
	"context"
	"fmt"
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
	"github.com/Nzyazin/zadnik.store/internal/gateway/admin"
	"github.com/Nzyazin/zadnik.store/internal/gateway/auth"
	"github.com/Nzyazin/zadnik.store/internal/gateway/middleware"
	admin_templates "github.com/Nzyazin/zadnik.store/internal/templates/admin-templates"
)

type ServerConfig struct {
	AuthServiceAddr string
	ProductServiceAddr string
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
	s.router.Use(middleware.PrometheusMiddleware())

	// Static files
	s.router.Static("/static", "./bin/static")
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Добавляем функции в шаблоны
	s.router.SetFuncMap(template.FuncMap{
		"staticWithHash": StaticWithHash,
	})

	// Подключаемся к auth сервису
	authConn, err := grpc.Dial(cfg.AuthServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	authClient := pb.NewAuthServiceClient(authConn)

	// Инициализация сервисов
	authService := NewAuthService(authClient)

	
	templates, err := admin_templates.NewTemplates(admin_templates.TemplateFunctions{
		StaticWithHash: StaticWithHash,
		Add: func(a, b int) int {
			return a + b
		},
	})
	if err != nil {
		return nil, err
	}

	productServiceUrl := fmt.Sprintf("http://%s", cfg.ProductServiceAddr)
	adminHandler := admin.NewHandler(authService, templates, productServiceUrl)
	adminHandler.RegisterRoutes(s.router)

	return s, nil
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

type authService struct {
	client pb.AuthServiceClient
}

func NewAuthService(client pb.AuthServiceClient) auth.AuthService {
	return &authService{client: client}
}

func (s *authService) Login(ctx context.Context, username, password string) (*pb.LoginResponse, error) {
	return s.client.Login(ctx, &pb.LoginRequest{Username: username, Password: password})
}

func (s *authService) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenResponse, error) {
	return s.client.ValidateToken(ctx, &pb.ValidateTokenRequest{AccessToken: token})
}