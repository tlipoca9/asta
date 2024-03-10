package fiberx

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.opentelemetry.io/otel/trace"
)

var (
	LoggerFormatConsole string
	LoggerFormatJSON    string
)

func init() {
	loggerFormatConsoleStrSlice := []string{
		"${time}",
		"${level}",
		"${msg}",
		"${request_id}",
		"${trace_id}",
		"${span_id}",
		"${status}",
		"${latency}",
		"${ip}",
		"${method} ${url}",
		"${error}",
	}

	const loggerFormatJSONStr = `
{
  "time": "${time}",
  "level": "${level}",
  "msg": "${msg}",
  "request_id": "${request_id}",
  "trace_id": "${trace_id}",
  "span_id": "${span_id}",
  "status": "${status}",
  "latency": "${latency}",
  "ip": "${ip}",
  "method": "${method}",
  "url": "${url}",
  "error": "${error}"
}
`
	LoggerFormatConsole = strings.Join(loggerFormatConsoleStrSlice, " | ") + "\n"

	var buf bytes.Buffer
	_ = json.Compact(&buf, []byte(loggerFormatJSONStr))
	buf.WriteByte('\n')
	LoggerFormatJSON = buf.String()
}

func LoggerTagLevel() logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		status := c.Response().StatusCode()
		lvl := slog.LevelInfo
		if status >= fiber.StatusInternalServerError {
			lvl = slog.LevelError
		} else if status >= fiber.StatusBadRequest {
			lvl = slog.LevelWarn
		}
		return output.WriteString(lvl.String())
	}
}

func LoggerTagMsg(s string) logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		return output.WriteString(s)
	}
}

func LoggerTagRequestID(key string) logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		return output.WriteString(fmt.Sprint(c.Locals(key)))
	}
}

func LoggerTagTraceID() logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		span := trace.SpanFromContext(c.UserContext())
		return output.WriteString(span.SpanContext().TraceID().String())
	}
}

func LoggerTagSpanID() logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		span := trace.SpanFromContext(c.UserContext())
		return output.WriteString(span.SpanContext().SpanID().String())
	}
}

func LoggerTagTagLatency() logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		latency := data.Stop.Sub(data.Start)
		return output.WriteString(latency.String())
	}
}

func LoggerConfigConsole(next func(*fiber.Ctx) bool) logger.Config {
	return logger.Config{
		Next: next,
		CustomTags: map[string]logger.LogFunc{
			"level":      LoggerTagLevel(),
			"msg":        LoggerTagMsg("access"),
			KeyRequestID: LoggerTagRequestID(KeyRequestID),
			KeyTraceID:   LoggerTagTraceID(),
			KeySpanID:    LoggerTagSpanID(),
		},
		Format: LoggerFormatConsole,
	}
}

func LoggerConfigJSON(next func(*fiber.Ctx) bool) logger.Config {
	return logger.Config{
		Next: next,
		CustomTags: map[string]logger.LogFunc{
			"level":      LoggerTagLevel(),
			"msg":        LoggerTagMsg("access"),
			"latency":    LoggerTagTagLatency(),
			KeyRequestID: LoggerTagRequestID(KeyRequestID),
			KeyTraceID:   LoggerTagTraceID(),
			KeySpanID:    LoggerTagSpanID(),
		},
		Format:        LoggerFormatJSON,
		TimeFormat:    time.RFC3339Nano,
		DisableColors: true,
	}
}
