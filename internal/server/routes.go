package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tlipoca9/asta/internal/handler"
)

func (s *FiberServer) RegisterFiberRoutes() {
	s.App.Get("/", s.HelloWorldHandler)

	s.App.Get("/health", s.healthHandler)

}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	h := handler.NewHealthHandler(
		handler.NewNamedHealther("database", s.db),
	)

	resp, _ := h.Health()

	return c.JSON(resp)
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	h := handler.NewHealthHandler(
		handler.NewNamedHealther("database", s.db),
	)

	resp, _ := h.Health()

	return c.JSON(resp)
}
