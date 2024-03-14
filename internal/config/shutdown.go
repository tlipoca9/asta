package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/tlipoca9/errors"

	"github.com/tlipoca9/asta/pkg/funcx"
)

var (
	defaultShutdownManager shutdownManager
	defaultShutdownNames   = make(map[string]struct{}, 0)
)

func DeferShutdown(name string, fn any) {
	// get caller
	var frame *runtime.Frame
	callers := make([]uintptr, 3)
	length := runtime.Callers(2, callers[:])
	callers = callers[:length]
	runtimeFrames := runtime.CallersFrames(callers)
	if f, more := runtimeFrames.Next(); more {
		frame = &f
	}
	if frame != nil {
		name = fmt.Sprintf("%s(%s:%d => %s())", name, frame.File, frame.Line, frame.Function)
	}
	// check if name already exists
	for _, ok := defaultShutdownNames[name]; ok; _, ok = defaultShutdownNames[name] {
		name = name + time.Now().Format("-20060102150405-") + ulid.Make().String()
	}
	defaultShutdownNames[name] = struct{}{}
	defaultShutdownManager.DeferShutdown(NewNamedShutdown(name, fn))
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

	maxTimeout := C.Service.ShutdownTimeout * time.Duration(len(defaultShutdownManager.shutdowns))
	ctx, cancel := context.WithTimeout(context.Background(), maxTimeout)
	defer cancel()
	log.Info(
		"start shutdown",
		slog.Duration("timeout", C.Service.ShutdownTimeout),
		slog.Duration("max_timeout", maxTimeout),
	)
	defaultShutdownManager.Shutdown(ctx, C.Service.ShutdownTimeout)
}

type shutdownManager struct {
	shutdowns []NamedShutdown
	mux       sync.Mutex
}

func (s *shutdownManager) DeferShutdown(sd NamedShutdown) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.shutdowns = append(s.shutdowns, sd)
}

func (s *shutdownManager) Shutdown(ctx context.Context, timeout time.Duration) {
	s.mux.Lock()
	defer s.mux.Unlock()
	// defer calls are executed in reverse order
	for i := len(s.shutdowns) - 1; i >= 0; i-- {
		sd := s.shutdowns[i]
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		if err := sd.Shutdown(timeoutCtx); err != nil {
			log.Error("shutdown failed", "name", sd.Name(), "error", err)
		} else {
			log.Info("shutdown success", "name", sd.Name())
		}
		cancel()
	}

	log.Info("shutdown complete")
}

type NamedShutdown interface {
	Name() string
	Shutdown(ctx context.Context) error
}

type NamedShutdownImpl struct {
	name string
	fn   any
}

func NewNamedShutdown(name string, fn any) NamedShutdown {
	return &NamedShutdownImpl{name: name, fn: fn}
}

func (s *NamedShutdownImpl) Name() string {
	return s.name
}

func (s *NamedShutdownImpl) Shutdown(ctx context.Context) error {
	var err error
	switch fn := s.fn.(type) {
	case func(context.Context) error:
		err = funcx.WrapContextE(ctx, fn)
	case func(context.Context):
		err = funcx.WrapContext(ctx, fn)
	case func() error:
		err = funcx.ContextE(ctx, fn)
	case func():
		err = funcx.Context(ctx, fn)
	default:
		err = errors.New("unknown shutdown function type")
	}
	return err
}
