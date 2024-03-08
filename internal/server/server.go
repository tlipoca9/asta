package server

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/tlipoca9/asta/internal/database"
)

type Server struct {
	*fiber.App

	log *slog.Logger
	db  database.Service
}

func New() *Server {
	server := &Server{
		App: fiber.New(),
		log: slog.Default(),
		db:  database.New(),
	}

	return server
}
