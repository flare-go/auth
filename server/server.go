package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"goflare.io/auth/authentication"
	"goflare.io/auth/authorization"
	"goflare.io/auth/handler"
	"goflare.io/auth/middleware"
)

// Server represents the server
type Server struct {
	mux            *http.ServeMux
	server         *http.Server
	authentication authentication.Service
	authorization  authorization.Service
	user           handler.UserHandler
	middleware     *middleware.AuthenticationMiddleware
	logger         *zap.Logger
}

// NewServer creates a new server
func NewServer(
	middleware *middleware.AuthenticationMiddleware,
	user handler.UserHandler,
	authentication authentication.Service,
	authorization authorization.Service,
	logger *zap.Logger,
) *Server {
	mux := http.NewServeMux()
	return &Server{
		mux:            mux,
		authentication: authentication,
		authorization:  authorization,
		middleware:     middleware,
		user:           user,
		logger:         logger,
	}
}

// Start starts the server
func (s *Server) Start(address string) error {
	s.registerRoutes()
	s.server = &http.Server{
		Addr:    address,
		Handler: s.mux,
	}
	return s.server.ListenAndServe()
}

func (s *Server) Run(address string) error {
	go func() {
		// Load policies
		if err := s.authorization.LoadPolicies(context.Background()); err != nil {
			s.logger.Fatal("Failed to load policies", zap.Error(err))
		}
		s.logger.Info("Starting server on " + address)
		if err := s.Start(address); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/login", s.user.Login)
	s.mux.HandleFunc("/register", s.user.Register)
	s.mux.HandleFunc("/check", s.middleware.AuthorizeUser(s.user.CheckPermission))
}
