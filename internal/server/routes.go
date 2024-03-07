package server

import (
	"log/slog"

	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/tlipoca9/asta/internal/handler"
	"github.com/tlipoca9/asta/pkg/logx"
	"go.opentelemetry.io/otel"
)

func (s *FiberServer) RegisterFiberMiddlewares() {
	s.App.Use(otelfiber.Middleware())
}

func (s *FiberServer) RegisterFiberRoutes() {
	s.RegisterFiberMiddlewares()

	s.App.Get("/", s.HelloWorldHandler())
	s.App.Get("/health", s.healthHandler)
}

func (s *FiberServer) HelloWorldHandler() fiber.Handler {
	tracer := otel.Tracer("hello-world")
	return func(c *fiber.Ctx) error {
		ctx, span := tracer.Start(c.UserContext(), "handler")
		defer span.End()
		ctx = logx.AddToContext(
			ctx,
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
		s.log.InfoContext(ctx, "start")
		defer s.log.InfoContext(ctx, "end")

		h := handler.NewHealthHandler(
			handler.NewNamedHealthier("database", s.db),
		)

		resp, _ := h.Health()

		return c.JSON(resp)
	}
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	h := handler.NewHealthHandler(
		handler.NewNamedHealthier("database", s.db),
	)

	resp, _ := h.Health()

	return c.JSON(resp)
}
