package server

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"

	"github.com/tlipoca9/asta/pkg/fiberx"
)

func (s *Server) initLogger(c *fiber.Ctx) *slog.Logger {
	span := trace.SpanFromContext(c.UserContext())
	log := s.log.With(
		slog.Any(fiberx.ContextKeyRequestID.String(), c.Locals(fiberx.ContextKeyRequestID)),
		slog.String(fiberx.ContextKeyTraceID.String(), span.SpanContext().TraceID().String()),
		slog.String(fiberx.ContextKeySpanID.String(), span.SpanContext().SpanID().String()),
	)
	return log
}
