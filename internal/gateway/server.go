package gateway

import (
	"fmt"
	"html/template"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
	"github.com/Nzyazin/zadnik.store/internal/gateway/admin"
	"github.com/Nzyazin/zadnik.store/internal/gateway/auth"
	"github.com/Nzyazin/zadnik.store/internal/gateway/middleware"
	"github.com/Nzyazin/zadnik.store/internal/broker"
	admin_templates "github.com/Nzyazin/zadnik.store/internal/templates/admin-templates"
)

type ServerConfig struct {
	AuthServiceAddr string
	ProductServiceAddr string
	ProductServiceAPIKey string
	Development    bool
	UserHTTPS      bool
	RabbitMQ broker.RabbitMQConfig
}

type Server struct {
	router *gin.Engine
	cfg    *ServerConfig
	messageBroker broker.MessageBroker
	httpServer *http.Server
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

	messageBroker, err := broker.NewRabbitMQBroker(
		broker.RabbitMQConfig{
			URL: cfg.RabbitMQ.URL,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize message broker: %w", err)
	}

	s.messageBroker = messageBroker

	// Подключаемся к auth сервису
	authConn, err := grpc.NewClient(cfg.AuthServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	authClient := pb.NewAuthServiceClient(authConn)

	// Инициализация сервисов
	authService := auth.NewGRPCAuthService(authClient)

	
	templates, err := admin_templates.NewTemplates(admin_templates.TemplateFunctions{
		StaticWithHash: StaticWithHash,
		Add: func(a, b int) int {
			return a + b
		},
	})
	if err != nil {
		return nil, err
	}

	protocol := "http"
	if cfg.UserHTTPS {
		protocol = "https"
	}

	productServiceUrl := fmt.Sprintf("%s://%s", protocol, cfg.ProductServiceAddr)
	adminHandler := admin.NewHandler(authService, templates, productServiceUrl, cfg.ProductServiceAPIKey, messageBroker)
	adminHandler.RegisterRoutes(s.router)

	return s, nil
}

func (s *Server) Run(addr string) error {
	srv := &http.Server{
		Addr: addr,
		Handler: s.router,
	}

	s.httpServer = srv

	return srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	var shutdownErr error
	go func() {
		if s.httpServer != nil {
			if err := s.httpServer.Shutdown(ctx); err != nil {
				shutdownErr = fmt.Errorf("HTTP server shutdown error: %w", err)
			}
		}

		if s.messageBroker != nil {
			if err := s.messageBroker.Close(); err != nil {
				if shutdownErr == nil {
					shutdownErr = fmt.Errorf("failed to close message broker: %w", err)
				} else {
					shutdownErr = fmt.Errorf("%w; also failed to close message broker: %v", shutdownErr, err)
				}				
			}
		}

		close(done)
	}()

	select {
	case <-done:
		return shutdownErr
	case <-ctx.Done():
		return fmt.Errorf("shutdown timed out: %w", ctx.Err())
	}
}
