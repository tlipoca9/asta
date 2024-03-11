package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"

	"github.com/DataDog/gostackparse"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/oklog/ulid/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/tlipoca9/asta/internal/config"
	"github.com/tlipoca9/asta/pkg/fiberx"
	"github.com/tlipoca9/asta/pkg/logx"

	_ "embed"
)

//go:embed favicon.ico
var faviconFile []byte

func (s *Server) RegisterMiddlewares() {
	commonNext := func(c *fiber.Ctx) bool {
		return c.Path() == "/healthz" || c.Path() == "/readyz" || strings.HasPrefix(c.Path(), "/debug")
	}

	// see https://docs.gofiber.io/api/middleware/recover
	if config.C.Service.Console {
		s.App.Use(recover.New(recover.Config{EnableStackTrace: true}))
	} else {
		s.App.Use(recover.New(recover.Config{
			EnableStackTrace: true,
			StackTraceHandler: func(_ *fiber.Ctx, e any) {
				goroutines, _ := gostackparse.Parse(bytes.NewReader(debug.Stack()))
				_, _ = fmt.Fprintln(os.Stderr, e)
				_ = json.NewEncoder(os.Stderr).Encode(goroutines)
			},
		}))
	}

	// see https://docs.gofiber.io/api/middleware/healthcheck
	s.App.Use(
		healthcheck.New(healthcheck.Config{
			LivenessProbe:     func(_ *fiber.Ctx) bool { return true },
			LivenessEndpoint:  "/healthz",
			ReadinessProbe:    func(_ *fiber.Ctx) bool { return s.db.Health() && s.cache.Health() },
			ReadinessEndpoint: "/readyz",
		}),
		otelfiber.Middleware(otelfiber.WithNext(commonNext)),
	)

	//// see https://docs.gofiber.io/api/middleware/requestid
	s.App.Use(requestid.New(requestid.Config{
		Header:     fiber.HeaderXRequestID,
		Generator:  func() string { return ulid.Make().String() },
		ContextKey: fiberx.ContextKeyRequestID,
	}))

	s.App.Use(func(c *fiber.Ctx) error {
		ctx, span := c.UserContext(), trace.SpanFromContext(c.UserContext())
		ctx = logx.AppendCtx(
			ctx,
			slog.Any(fiberx.ContextKeyRequestID.String(), c.Locals(fiberx.ContextKeyRequestID)),
			slog.String(fiberx.ContextKeyTraceID.String(), span.SpanContext().TraceID().String()),
			slog.String(fiberx.ContextKeySpanID.String(), span.SpanContext().SpanID().String()),
		)
		c.SetUserContext(ctx)
		return c.Next()
	})

	if config.C.Service.Debug {
		// see https://docs.gofiber.io/api/middleware/monitor
		s.App.Get("/debug/metrics/ui", monitor.New())
		// see https://docs.gofiber.io/api/middleware/pprof
		s.App.Use(pprof.New())
	}

	// see https://prometheus.io/docs/guides/go-application
	s.App.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	// see https://docs.gofiber.io/api/middleware/logger
	if config.C.Service.Console {
		s.App.Use(logger.New(fiberx.LoggerConfigConsole(commonNext)))
	} else {
		s.App.Use(logger.New(fiberx.LoggerConfigJSON(commonNext)))
	}

	// see https://docs.gofiber.io/api/middleware/favicon
	s.App.Use(favicon.New(favicon.Config{
		Data: faviconFile,
		URL:  "/favicon.ico",
	}))
}

func (s *Server) RegisterRoutes() {
	s.App.Get("/", s.HelloWorldHandler())
}

func (s *Server) HelloWorldHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.UserContext()

		s.log.InfoContext(ctx, "start")
		defer s.log.InfoContext(ctx, "end")

		return c.JSON(fiber.Map{"message": "Hello, World!", "key": c.OriginalURL(), "query": c.Queries()})
	}
}
