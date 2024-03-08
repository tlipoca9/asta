package server

import (
	"log/slog"

	fiber "github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"

	"github.com/tlipoca9/asta/pkg/fiberx"
)

func (s *Server) initLogger(c *fiber.Ctx) *slog.Logger {
	span := trace.SpanFromContext(c.UserContext())
	log := s.log.With(
		slog.Any(fiberx.KeyRequestID, c.Locals(fiberx.KeyRequestID)),
		slog.String(fiberx.KeyTraceID, span.SpanContext().TraceID().String()),
		slog.String(fiberx.KeySpanID, span.SpanContext().SpanID().String()),
	)
	return log
}
