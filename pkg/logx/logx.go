package logx

// see https://github.com/golang/go/issues/56345#issuecomment-1435247981

import (
	"context"
	"log/slog"
)

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
