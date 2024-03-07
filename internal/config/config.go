package config

import (
	"context"
	validator "github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	koanf "github.com/knadh/koanf/v2"
	"github.com/tlipoca9/asta/pkg/logx"
	"go.opentelemetry.io/otel"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"log/slog"
	"os"
)

var (
	log = slog.New(logx.NewHandler(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	)))
	validate = validator.New(
		validator.WithRequiredStructEnabled(),
	)
	tp *sdktrace.TracerProvider
	C  Config
)

func initTracer() {
	err := os.MkdirAll("run", 0750)
	if err != nil {
		panic(err)
	}
	out, err := os.OpenFile("run/trace.log", os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	exporter, err := stdout.New(stdout.WithWriter(out))
	if err != nil {
		panic(err)
	}
	tp = sdktrace.NewTracerProvider(
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
}

func Shutdown(ctx context.Context) {
	if err := tp.Shutdown(ctx); err != nil {
		log.Error("shutting down tracer provider", "error", err)
	}
}

type Config struct {
	ServiceName string `json:"service_name" validate:"required"`
}

func init() {
	slog.SetDefault(log)
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

	initTracer()
}
