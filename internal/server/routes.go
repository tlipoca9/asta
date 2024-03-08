package server

import (
	"log/slog"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"go.opentelemetry.io/otel"
)

func (s *Server) RegisterMiddlewares() {
	s.App.Use(
		healthcheck.New(healthcheck.Config{
			LivenessProbe:     func(c *fiber.Ctx) bool { return true },
			LivenessEndpoint:  "/healthz",
			ReadinessProbe:    func(ctx *fiber.Ctx) bool { return s.db.Health() },
			ReadinessEndpoint: "/readyz",
		}),
		otelfiber.Middleware(),
	)
}

func (s *Server) RegisterRoutes() {
	s.RegisterMiddlewares()

	s.App.Get("/", s.HelloWorldHandler())
}

func (s *Server) HelloWorldHandler() fiber.Handler {
	tracer := otel.Tracer("hello-world")
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.UserContext(), "handler")
		defer span.End()
		log := s.log.With(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		log.InfoContext(ctx, "start")
		defer log.InfoContext(ctx, "end")

		return c.JSON(fiber.Map{"message": "Hello, World!"})
	}
}
