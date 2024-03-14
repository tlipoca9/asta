package server

import (
	"log/slog"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/tlipoca9/errors"

	"github.com/tlipoca9/asta/internal/cache"
	"github.com/tlipoca9/asta/internal/config"
	"github.com/tlipoca9/asta/internal/database"
)

func Serve() error {
	s := newServer()
	s.RegisterMiddlewares()
	s.RegisterRoutes()
	config.DeferShutdown("server", s.ShutdownWithContext)
	return errors.Wrap(s.Serve(config.C.Service.Addr), "start server failed")
}

type Server struct {
	*fiber.App

	log   *slog.Logger
	db    database.Service
	cache cache.Service
}

func newServer() *Server {
	server := &Server{
		App: fiber.New(fiber.Config{
			JSONEncoder: json.Marshal,
			JSONDecoder: json.Unmarshal,
		}),
		log:   slog.Default(),
		db:    database.New(),
		cache: cache.New(),
	}

	return server
}

func (s *Server) Serve(addr string) error {
	s.RegisterMiddlewares()
	s.RegisterRoutes()
	return s.Listen(addr)
}
