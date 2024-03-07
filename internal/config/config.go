package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/tlipoca9/asta/pkg/funcx"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Config struct {
	ServiceName     string        `json:"service_name,omitempty" validate:"required"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout,omitempty" validate:"required"`

	Database struct {
		Enable   bool   `json:"enable,omitempty"`
		DBName   string `json:"db_name,omitempty" validate:"required_if=Enable true"`
		Password string `json:"password,omitempty" validate:"required_if=Enable true"`
		Username string `json:"username,omitempty" validate:"required_if=Enable true"`
		Port     int    `json:"port,omitempty" validate:"required_if=Enable true"`
		Host     string `json:"host,omitempty" validate:"required_if=Enable true"`
	}
}

var (
	log       *slog.Logger
	shutdowns sync.Map
	validate  = validator.New(validator.WithRequiredStructEnabled())
	C         Config
)

func initLogger() {
	log = slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	))
	slog.SetDefault(log)
}

func initConfig() {
	k := koanf.New(".")
	if err := k.Load(file.Provider("etc/config.toml"), toml.Parser()); err != nil {
		panic(err)
	}

	err := k.UnmarshalWithConf("", nil, koanf.UnmarshalConf{
		Tag: "json",
		DecoderConfig: &mapstructure.DecoderConfig{
			Result:  &C,
			TagName: "json",
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.TextUnmarshallerHookFunc(),
			),
		},
	})
	if err != nil {
		panic(err)
	}

	err = validate.Struct(C)
	if err != nil {
		panic(err)
	}

	log.Info("service config", "config", C)
}

func initTracer() {
	err := os.MkdirAll("run", 0o750)
	if err != nil {
		panic(err)
	}
	out, err := os.OpenFile("run/trace.log", os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}
	exporter, err := stdout.New(stdout.WithWriter(out))
	if err != nil {
		panic(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(C.ServiceName),
			)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	RegisterShutdown("tracer-provider", tp.Shutdown)
}

func RegisterShutdown(name string, fn any) {
	_, found := shutdowns.Load(name)
	for found {
		name = fmt.Sprintf("%s-%d", name, time.Now().Unix())
		_, found = shutdowns.Load(name)
	}
	shutdowns.Store(name, fn)
}

func WaitForExit() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

		s := <-ch
		log.Info("catch exit signal", slog.String("signal", s.String()))

		ctx, cancel := context.WithTimeout(context.Background(), C.ShutdownTimeout)
		defer cancel()

		var wgg sync.WaitGroup
		shutdowns.Range(func(k, v any) bool {
			wgg.Add(1)
			go func() {
				defer wgg.Done()
				var err error
				log := log.With("name", k)
				switch fn := v.(type) {
				case func(context.Context) error:
					err = funcx.WrapContextE(ctx, fn)
				case func(context.Context):
					err = funcx.WrapContext(ctx, fn)
				case func() error:
					err = funcx.ContextE(ctx, fn)
				case func():
					err = funcx.Context(ctx, fn)
				default:
					log.Warn("unknown shutdown function type, skip")
				}
				if err != nil {
					log.Error("shutdown failed", "error", err)
				} else {
					log.Info("shutdown success")
				}
			}()
			return true
		})
		wgg.Wait()
	}()
	wg.Wait()
}

func init() {
	initLogger()
	initConfig()
	initTracer()
}
