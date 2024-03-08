package server

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/oklog/ulid/v2"

	"github.com/DataDog/gostackparse"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/tlipoca9/asta/internal/config"
)

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
		ContextKey: KeyRequestID,
	}))

	if config.C.Service.Debug {
		// see https://docs.gofiber.io/api/middleware/monitor
		s.App.Get("/debug/metrics/ui", monitor.New())
		// see https://docs.gofiber.io/api/middleware/pprof
		s.App.Use(pprof.New())
	}

	s.App.Use(func(c *fiber.Ctx) error {
		if commonNext(c) {
			return c.Next()
		}

		start := time.Now()

		err := c.Next()

		ctx, log := c.UserContext(), s.initLogger(c)
		lvl := slog.LevelDebug
		if err != nil {
			lvl = slog.LevelError
			log = log.With("error", err)
		}

		latency := time.Since(start)
		ip := c.IP()
		method := c.Method()
		path := c.OriginalURL()
		status := c.Response().StatusCode()

		log.Log(ctx, lvl, "access", "latency", latency, "ip", ip, "status", status, "method", method, "path", path)

		return err
	})
	// see https://docs.gofiber.io/api/middleware/logger
	//if config.C.Service.Console {
	//	s.App.Use(logger.New(logger.Config{Next: commonNext}))
	//} else {
	//	s.App.Use(logger.New(logger.Config{
	//		Next: commonNext,
	//		// json format
	//		Format:        "{\"time\":\"${time}\",\"status\":\"${status}\",\"latency\":\"${latency}\",\"ip\":\"${ip}\",\"method\":\"${method}\",\"path\":\"${path}\",\"error\":\"${error}\"}\n",
	//		TimeFormat:    time.RFC3339,
	//		Output:        os.Stdout,
	//		DisableColors: true,
	//	}))
	//}
}

func (s *Server) RegisterRoutes() {
	s.RegisterMiddlewares()

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
