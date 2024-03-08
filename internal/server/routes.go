package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/DataDog/gostackparse"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/tlipoca9/asta/internal/config"
	"go.opentelemetry.io/otel"
)

func (s *Server) RegisterMiddlewares() {
	commonNext := func(c *fiber.Ctx) bool {
		return c.Path() == "/healthz" || c.Path() == "/readyz" || strings.HasPrefix(c.Path(), "/debug")
	}

	if config.C.Service.Console {
		s.App.Use(recover.New(recover.Config{EnableStackTrace: true}))
	} else {
		s.App.Use(recover.New(recover.Config{
			EnableStackTrace: true,
			StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
				goroutines, _ := gostackparse.Parse(bytes.NewReader(debug.Stack()))
				_ = json.NewEncoder(os.Stderr).Encode(goroutines)
			},
		}))
	}

	s.App.Use(
		healthcheck.New(healthcheck.Config{
			LivenessProbe:     func(c *fiber.Ctx) bool { return true },
			LivenessEndpoint:  "/healthz",
			ReadinessProbe:    func(ctx *fiber.Ctx) bool { return s.db.Health() },
			ReadinessEndpoint: "/readyz",
		}),
		otelfiber.Middleware(otelfiber.WithNext(commonNext)),
	)

	if config.C.Service.Console {
		s.App.Use(logger.New(logger.Config{Next: commonNext}))
	} else {
		s.App.Use(logger.New(logger.Config{
			Next: commonNext,
			// json format
			Format:        "{\"time\":\"${time}\",\"status\":\"${status}\",\"latency\":\"${latency}\",\"ip\":\"${ip}\",\"method\":\"${method}\",\"path\":\"${path}\",\"error\":\"${error}\"}\n",
			TimeFormat:    time.RFC3339,
			Output:        os.Stdout,
			DisableColors: true,
		}))
	}

	if config.C.Service.Debug {
		s.App.Get("/debug/metrics/ui", monitor.New())
		s.App.Use(pprof.New())
	}
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
		panic("hh")

		return c.JSON(fiber.Map{"message": "Hello, World!", "key": c.OriginalURL(), "query": c.Queries()})
	}
}
