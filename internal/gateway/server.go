package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Nzyazin/zadnik.store/api/generated/auth"
	common "github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/Nzyazin/zadnik.store/internal/broker"
	"github.com/Nzyazin/zadnik.store/internal/gateway/admin"
	"github.com/Nzyazin/zadnik.store/internal/gateway/client"
	"github.com/Nzyazin/zadnik.store/internal/gateway/auth"
	"github.com/Nzyazin/zadnik.store/internal/gateway/middleware"
	admin_templates "github.com/Nzyazin/zadnik.store/internal/templates/admin-templates"
	client_templates "github.com/Nzyazin/zadnik.store/internal/templates/client-templates"
)

type SMTPConfig struct {
	Host string
	Port int
	From string
	Password string
}

type ServerConfig struct {
	AuthServiceAddr string
	ProductServiceAddr string
	ProductServiceAPIKey string
	Development    bool
	UseHTTPS      bool
	RabbitMQ broker.RabbitMQConfig
	SMTPConfig SMTPConfig
	CertFile string
	KeyFile string
	LOG_FILE string
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

	logger := common.NewSimpleLogger(&common.LogConfig{FilePath: cfg.LOG_FILE})
	s.router.Use(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()
		timestamp := time.Now()
		latency := timestamp.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logMessage := fmt.Sprintf("%s | %3d | %13v | %15s | %-7s %s",
			timestamp.Format("2006-01-02 15:04:05"),
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)
		logger.Infof(logMessage)
	})


	// Middleware
	s.router.Use(gin.Recovery())
	s.router.Use(middleware.PrometheusMiddleware())

	s.router.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/storage/images") || strings.HasPrefix(path, "/static"){
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		c.Next()
	})

	// Static files
	s.router.Static("/static/admin", "./bin/static/admin")
	s.router.Static("/static/client", "./bin/static/client")
	s.router.Static("/storage/images", "./storage/images")
	s.router.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./bin/static/admin/images/favicon/favicon.ico")
	})
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	messageBroker, err := broker.NewRabbitMQBroker(
		broker.RabbitMQConfig{
			URL: cfg.RabbitMQ.URL,
			LogFilePath: cfg.LOG_FILE,
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

	
	adminTemplates, err := admin_templates.NewTemplates(admin_templates.TemplateFunctions{
		StaticWithHash: StaticWithHash,
		Add: Add,
		Dict: Dict,
	})
	if err != nil {
		return nil, err
	}

	clientTemplates, err := client_templates.NewTemplates(client_templates.TemplateFunctions{
		StaticWithHash: StaticWithHash,
	})
	if err != nil {
		return nil, err
	}

	protocol := "http"
	if cfg.UseHTTPS {
		protocol = "https"
	}

	emailSender := client.NewSMTPEmailSender(
		cfg.SMTPConfig.Host,
		cfg.SMTPConfig.Port,
		cfg.SMTPConfig.From,
		cfg.SMTPConfig.Password,
		common.NewSimpleLogger(&common.LogConfig{FilePath: cfg.LOG_FILE}),
	)

	productServiceUrl := fmt.Sprintf("%s://%s", protocol, cfg.ProductServiceAddr)
	adminHandler := admin.NewHandler(authService, adminTemplates, productServiceUrl, cfg.ProductServiceAPIKey, messageBroker)
	clientHandler := client.NewHandler(clientTemplates, productServiceUrl, cfg.ProductServiceAPIKey, emailSender)
	clientHandler.RegisterRoutes(s.router)
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

func (s *Server) RunWithTLS(addr, certFile, keyFile string) error {
	srv := &http.Server{
		Addr: addr,
		Handler: s.router,
	}

	s.httpServer = srv

	return srv.ListenAndServeTLS(certFile, keyFile)
}

func RunHTTPRedirect(httpAddr, httpsHost string) *http.Server {
	if strings.HasPrefix(httpsHost, "http://") {
		httpsHost = httpsHost[len("http://"):]
	} else if strings.HasPrefix(httpsHost, "https://") {
		httpsHost = httpsHost[len("https://"):]
	}
	srv := &http.Server{
		Addr: httpAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + httpsHost + r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}),
	}

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("HTTP redirect server error: %v", err)
		}
	}()

	return srv
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
