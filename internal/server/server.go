package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tlipoca9/asta/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(),
		db: database.New(database.Config{
			DBName:   "asta",
			Password: "asta",
			Username: "asta",
			Port:     "3306",
			Host:     "localhost",
		}),
	}

	return server
}
