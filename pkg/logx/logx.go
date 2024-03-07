package logx

// see https://github.com/golang/go/issues/56345#issuecomment-1435247981

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*ContextHandler)(nil)

type ContextHandler struct {
	h slog.Handler
}

func (c ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return c.h.Enabled(ctx, level)
}

func (c ContextHandler) Handle(ctx context.Context, record slog.Record) error {
	record.AddAttrs(ContextAttrs(ctx)...)
	return c.h.Handle(ctx, record)
}

func (c ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c.h = c.h.WithAttrs(attrs)
	return c
}

func (c ContextHandler) WithGroup(name string) slog.Handler {
	c.h = c.h.WithGroup(name)
	return c
}

func NewHandler(handler slog.Handler) slog.Handler {
	return ContextHandler{
		h: handler,
	}
}

type ctxKey int

const argsKey ctxKey = 1

func AddToContext(ctx context.Context, args ...slog.Attr) context.Context {
	if len(args) == 0 {
		return ctx
	}
	attrs := ContextAttrs(ctx)
	attrs = append(attrs, args...)
	return context.WithValue(ctx, argsKey, attrs)
}

func ContextAttrs(ctx context.Context) []slog.Attr {
	attrs, ok := ctx.Value(argsKey).([]slog.Attr)
	if !ok {
		return nil
	}
	return attrs
}
