package server

import (
	"log/slog"

	"github.com/tlipoca9/asta/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/tlipoca9/asta/internal/database"
)

type Server struct {
	*fiber.App

	log *slog.Logger
	db  database.Service
}

func New() *Server {
	var db database.Service
	if config.C.Database.Enable {
		db = database.New(
			slog.Default(),
			database.Config{
				DBName:   config.C.Database.DBName,
				Password: config.C.Database.Password,
				Username: config.C.Database.Username,
				Port:     config.C.Database.Port,
				Host:     config.C.Database.Host,
			},
		)
	}

	server := &Server{
		App: fiber.New(),
		log: slog.Default(),
		db:  db,
	}

	return server
}
