package server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"goflare.io/auth/authentication"
	"goflare.io/auth/authorization"
	"goflare.io/auth/handler"
	authMiddleware "goflare.io/auth/middleware"
)

type Server struct {
	echo           *echo.Echo
	Authentication authentication.Service
	Authorization  authorization.Service
	User           handler.UserHandler
	Middleware     *authMiddleware.AuthenticationMiddleware
}

func NewServer(
	middleware *authMiddleware.AuthenticationMiddleware,
	User handler.UserHandler,
	Authentication authentication.Service,
	Authorization authorization.Service,
) *Server {
	return &Server{
		echo:           echo.New(),
		Authentication: Authentication,
		Authorization:  Authorization,
		Middleware:     middleware,
		User:           User,
	}
}

// Start initializes the server by registering middlewares and routes, and starts listening for connections on the provided address.
// It returns an error if there is an issue starting the server.
func (s *Server) Start(address string) error {
	s.registerMiddlewares()
	s.registerRoutes()
	return s.echo.Start(address)
}

// Run starts the server by calling the Start method in a goroutine. If an error occurs, it
// logs the error and terminates the server. It then listens for an OS interrupt signal or a SIGTERM
// signal to gracefully shut down the server. Once the signal is received, it creates a context with
// a timeout of 5 seconds, cancels the context after the method returns, and returns the result of
// shutting down the server.
func (s *Server) Run(address string) error {

	go func() {
		if err := s.Start(address); err != nil {
			s.echo.Logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.echo.Shutdown(ctx)
}

func (s *Server) registerMiddlewares() {
	s.echo.Use(middleware.Recover())
}

func (s *Server) registerRoutes() {

	s.echo.POST("/login", s.User.Login)
	s.echo.POST("/register", s.User.Register)

	s.echo.POST("/check", s.User.CheckPermission, s.Middleware.AuthorizeUser)
}
