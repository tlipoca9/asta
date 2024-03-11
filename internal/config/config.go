package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/tlipoca9/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/tlipoca9/asta/pkg/funcx"
	"github.com/tlipoca9/asta/pkg/logx"
)

type Config struct {
	Service struct {
		Name            string        `json:"name"`
		Addr            string        `json:"addr"`
		ShutdownTimeout time.Duration `json:"shutdown_timeout"`
		Console         bool          `json:"console"`
		Debug           bool          `json:"debug"`
	} `json:"service"`

	Database struct {
		DBName   string `json:"db_name"`
		Username string `json:"username"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	} `json:"database"`

	Cache struct {
		Address string `json:"address"`
	}
}

var (
	log       *slog.Logger
	shutdowns sync.Map
	C         Config
)

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

	b, _ := json.Marshal(C)
	fmt.Printf("config: %s\n", string(b))
}

func initErrors() {
	errors.C.Style = errors.StyleStack
	errors.C.StackFramesHandler = errors.JSONStackFramesHandler
}

func initLogger() {
	var (
		h        slog.Handler
		lvl      slog.Level
		replacer = func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == "error" || a.Key == "err" {
				return logx.JSON(a.Key, a.Value.Any())
			}

			if v, ok := a.Value.Any().(map[string]any); ok {
				return logx.JSON(a.Key, v)
			}

			if v, ok := a.Value.Any().([]map[string]any); ok {
				return logx.JSON(a.Key, v)
			}

			if v, ok := a.Value.Any().([]any); ok {
				return logx.JSON(a.Key, v)
			}

			return a
		}
	)
	if C.Service.Debug {
		lvl = slog.LevelDebug
	} else {
		lvl = slog.LevelInfo
	}
	if C.Service.Console {
		h = tint.NewHandler(colorable.NewColorableStderr(), &tint.Options{
			Level:       lvl,
			TimeFormat:  time.TimeOnly,
			ReplaceAttr: replacer,
		})
	} else {
		h = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:       lvl,
			ReplaceAttr: replacer,
		})
	}
	log = slog.New(logx.NewContextHandler(h))
	slog.SetDefault(log)
}

func initTracer() {
	err := os.MkdirAll("run", 0o750)
	if err != nil {
		panic(err)
	}
	out, err := os.OpenFile("run/trace.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		panic(err)
	}
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(out))
	if err != nil {
		panic(err)
	}
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithBatcher(exporter),
		trace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(C.Service.Name),
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

func WaitForExit(ctx context.Context) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		log.Info("context done, exit")
	case s := <-ch:
		log.Info("catch exit signal", slog.String("signal", s.String()))
	}
	ctx, cancel := context.WithTimeout(context.Background(), C.Service.ShutdownTimeout)
	defer cancel()

	var wg sync.WaitGroup
	shutdowns.Range(func(k, v any) bool {
		wg.Add(1)
		go func() {
			defer wg.Done()
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
	wg.Wait()
}

func init() {
	initConfig()
	initErrors()
	initLogger()
	initTracer()
}
