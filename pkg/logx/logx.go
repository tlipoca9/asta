package logx

// see https://github.com/golang/go/issues/56345#issuecomment-1435247981

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/goccy/go-json"
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
