package server

import (
	"log/slog"

	"github.com/tlipoca9/asta/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/tlipoca9/asta/internal/database"
)

type FiberServer struct {
	*fiber.App

	log *slog.Logger
	db  database.Service
}

func New() *FiberServer {
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

	server := &FiberServer{
		App: fiber.New(),
		log: slog.Default(),
		db:  db,
	}

	return server
}
