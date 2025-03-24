package server

import (
	"net/http"
	"context"

	"github.com/Nzyazin/zadnik.store/internal/common"
	"github.com/gorilla/mux"
	"github.com/Nzyazin/zadnik.store/internal/product/delivery"
)

type Server struct {
	srv *http.Server
	logger common.Logger
}

func NewServer(addr string, handler *delivery.ProductHandler, logger common.Logger) *Server {
	router := mux.NewRouter()
	router.Use(handler.AuthMiddleware)
	router.HandleFunc("/products", handler.GetAll).Methods("GET")
	router.HandleFunc("/products/{id}", handler.GetByID).Methods("GET")
	router.HandleFunc("/products/{id}", handler.Update).Methods("PATCH")

	return &Server{
		srv: &http.Server{
			Addr: addr,
			Handler: router,
		},
		logger: logger,
	}
}

func (s *Server) Run() error {
	s.logger.Infof("Starting product service on %s", s.srv.Addr)
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}