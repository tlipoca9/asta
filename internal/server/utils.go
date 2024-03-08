package server

import (
	"log/slog"

	fiber "github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/trace"
)

const (
	KeyRequestID = "request_id"
	KeyTraceID   = "trace_id"
	KeySpanID    = "span_id"
)

func (s *Server) initLogger(c *fiber.Ctx) *slog.Logger {
	span := trace.SpanFromContext(c.UserContext())
	log := s.log.With(
		slog.Any(KeyRequestID, c.Locals(KeyRequestID)),
		slog.String(KeyTraceID, span.SpanContext().TraceID().String()),
		slog.String(KeySpanID, span.SpanContext().SpanID().String()),
	)
	return log
}
