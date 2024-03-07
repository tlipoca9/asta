package server

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/tlipoca9/asta/internal/database"
)

type FiberServer struct {
	*fiber.App

	log *slog.Logger
	db  database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(),
		log: slog.Default(),
		//db: database.New(
		//	log,
		//	database.Config{
		//		DBName:   "asta",
		//		Password: "asta",
		//		Username: "asta",
		//		Port:     "3306",
		//		Host:     "localhost",
		//	},
		//),
	}

	return server
}
