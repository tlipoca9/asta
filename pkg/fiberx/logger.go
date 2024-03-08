package fiberx

import (
	"bytes"
	"fmt"
	"log/slog"

	json "github.com/goccy/go-json"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.opentelemetry.io/otel/trace"
)

const LoggerConsoleFormat = "${time} | ${level} | ${msg} | ${request_id} | ${trace_id} | ${span_id} | ${status} | ${latency} | ${ip} | ${method} ${url} | ${error}\n"

const loggerJSONFormatStr = `
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

var LoggerJSONFormat string

func init() {
	var buf bytes.Buffer
	_ = json.Compact(&buf, []byte(loggerJSONFormatStr))
	buf.WriteByte('\n')
	LoggerJSONFormat = buf.String()
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
