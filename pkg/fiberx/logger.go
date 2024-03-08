package fiberx

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"time"

	json "github.com/goccy/go-json"
	fiber "github.com/gofiber/fiber/v2"
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

func LoggerLevelTag() logger.LogFunc {
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

func LoggerMsgTag(s string) logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		return output.WriteString(s)
	}
}

func LoggerRequestIDTag(key string) logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		return output.WriteString(fmt.Sprint(c.Locals(key)))
	}
}

func LoggerTraceIDTag() logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		span := trace.SpanFromContext(c.UserContext())
		return output.WriteString(span.SpanContext().TraceID().String())
	}
}

func LoggerSpanIDTag() logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		span := trace.SpanFromContext(c.UserContext())
		return output.WriteString(span.SpanContext().SpanID().String())
	}
}

func LoggerLatencyTag() logger.LogFunc {
	return func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
		latency := data.Stop.Sub(data.Start)
		return output.WriteString(latency.String())
	}
}

func LoggerConfigConsole(next func(*fiber.Ctx) bool) logger.Config {
	return logger.Config{
		Next: next,
		CustomTags: map[string]logger.LogFunc{
			"level":      LoggerLevelTag(),
			"msg":        LoggerMsgTag("access"),
			KeyRequestID: LoggerRequestIDTag(KeyRequestID),
			KeyTraceID:   LoggerTraceIDTag(),
			KeySpanID:    LoggerSpanIDTag(),
		},
		Format: LoggerFormatConsole,
	}
}

func LoggerConfigJSON(next func(*fiber.Ctx) bool) logger.Config {
	return logger.Config{
		Next: next,
		CustomTags: map[string]logger.LogFunc{
			"level":      LoggerLevelTag(),
			"msg":        LoggerMsgTag("access"),
			"latency":    LoggerLatencyTag(),
			KeyRequestID: LoggerRequestIDTag(KeyRequestID),
			KeyTraceID:   LoggerTraceIDTag(),
			KeySpanID:    LoggerSpanIDTag(),
		},
		Format:        LoggerFormatJSON,
		TimeFormat:    time.RFC3339Nano,
		DisableColors: true,
	}
}
