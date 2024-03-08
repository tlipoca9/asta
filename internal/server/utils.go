package server

import (
	"log/slog"

	fiber "github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

const (
	KeyRequestID = "request_id"
)

func (s *Server) initLogger(c *fiber.Ctx) *slog.Logger {
	span := trace.SpanFromContext(c.UserContext())
	log := s.log.With(
		slog.Any(KeyRequestID, c.Locals(KeyRequestID)),
		slog.String("trace_id", span.SpanContext().TraceID().String()),
		slog.String("span_id", span.SpanContext().SpanID().String()),
	)
	return log
}
