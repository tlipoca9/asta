package logx

// see https://github.com/golang/go/issues/56345#issuecomment-1435247981

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/goccy/go-json"
)

type ContextKey int

//go:generate stringer -type=ContextKey -output=contextkey.gen.go -linecomment
const (
	ContextKeyAttrs ContextKey = iota + 1 // attrs
)

type ContextHandler struct {
	slog.Handler
}

func NewContextHandler(h slog.Handler) slog.Handler {
	return ContextHandler{Handler: h}
}

// Handle adds contextual attributes to the Record before calling the underlying
// handler
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs, ok := ctx.Value(ContextKeyAttrs).([]slog.Attr); ok {
		r.AddAttrs(attrs...)
	}

	return h.Handler.Handle(ctx, r)
}

// AppendCtx adds a slog attribute to the provided context so that it will be
// included in any Record created with such context
func AppendCtx(parent context.Context, attr ...slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(ContextKeyAttrs).([]slog.Attr); ok {
		return context.WithValue(parent, ContextKeyAttrs, append(v, attr...))
	}

	return context.WithValue(parent, ContextKeyAttrs, attr)
}

func JSON(key string, val any) slog.Attr {
	ret := make([]map[string]any, 0)
	if v, ok := val.([]map[string]any); ok {
		ret = v
	} else {
		var buf []byte
		switch v := val.(type) {
		case string:
			buf = []byte(v)
		case []byte:
			buf = v
		case error:
			buf = []byte(v.Error())
		default:
			return slog.Any(key, val)
		}
		if err := json.Unmarshal(buf, &ret); err != nil {
			return slog.Any(key, val)
		}
	}

	var collect func(string, any) slog.Attr
	collect = func(key string, val any) slog.Attr {
		switch vv := val.(type) {
		case []any:
			attrs := make([]any, 0, len(vv))
			for i, v := range vv {
				attrs = append(attrs, collect(fmt.Sprint(i), v))
			}
			return slog.Group(key, attrs...)
		case []map[string]any:
			attrs := make([]any, 0, len(vv))
			for i, v := range vv {
				attrs = append(attrs, collect(fmt.Sprint(i), v))
			}
			return slog.Group(key, attrs...)
		case map[string]any:
			attrs := make([]any, 0, len(vv))
			for k, v := range vv {
				attrs = append(attrs, collect(k, v))
			}
			return slog.Group(key, attrs...)
		default:
			return slog.Any(key, val)
		}
	}
	return collect(key, ret)
}
