package gateway

import (
	"context"
	"html/template"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Nzyazin/zadnik.store/internal/gateway/auth" 
	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
	"github.com/Nzyazin/zadnik.store/internal/gateway/admin"
	admin_templates "github.com/Nzyazin/zadnik.store/internal/templates/admin-templates"
	"github.com/Nzyazin/zadnik.store/internal/gateway/middleware"
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
	})
	if err != nil {
		return nil, err
	}
	adminHandler := admin.NewHandler(authService, templates)
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
