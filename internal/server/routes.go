package server

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"runtime/debug"
	"strings"

	"github.com/DataDog/gostackparse"
	"github.com/gofiber/contrib/otelfiber"
	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	ulid "github.com/oklog/ulid/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/tlipoca9/asta/internal/config"
	"github.com/tlipoca9/asta/pkg/fiberx"
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
			StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
				goroutines, _ := gostackparse.Parse(bytes.NewReader(debug.Stack()))
				_ = json.NewEncoder(os.Stderr).Encode(goroutines)
			},
		}))
	}

	// see https://docs.gofiber.io/api/middleware/healthcheck
	s.App.Use(
		healthcheck.New(healthcheck.Config{
			LivenessProbe:     func(c *fiber.Ctx) bool { return true },
			LivenessEndpoint:  "/healthz",
			ReadinessProbe:    func(ctx *fiber.Ctx) bool { return s.db.Health() },
			ReadinessEndpoint: "/readyz",
		}),
		otelfiber.Middleware(otelfiber.WithNext(commonNext)),
	)

	//// see https://docs.gofiber.io/api/middleware/requestid
	s.App.Use(requestid.New(requestid.Config{
		Header:     fiber.HeaderXRequestID,
		Generator:  func() string { return ulid.Make().String() },
		ContextKey: fiberx.KeyRequestID,
	}))

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
		ctx, log := c.UserContext(), s.initLogger(c)

		log.InfoContext(ctx, "start")
		defer log.InfoContext(ctx, "end")

		return c.JSON(fiber.Map{"message": "Hello, World!", "key": c.OriginalURL(), "query": c.Queries()})
	}
}
