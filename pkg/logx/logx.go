package logx

import (
	"context"
	"log/slog"
)

type ctxKey int

const loggerKey ctxKey = 1

type loggerHook func(context.Context, *slog.Logger) *slog.Logger

var hooks map[string]loggerHook

func RegisterHook(name string, hook loggerHook) {
	if hook == nil {
		panic("logx: RegisterHook hook is nil")
	}

	if hooks == nil {
		hooks = make(map[string]loggerHook)
	}
	hooks[name] = hook
}

func New(ctx ...context.Context) *slog.Logger {
	if len(ctx) > 1 {
		panic("logx: New received more than one context")
	}

	ctx0 := context.Background()
	if len(ctx) == 1 {
		ctx0 = ctx[0]
	}

	l := Context(ctx0)
	for _, hook := range hooks {
		l = hook(ctx0, l)
	}
	return l
}

func Context(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}
	return l
}

func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}
