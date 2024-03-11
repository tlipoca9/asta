package fiberx

type ContextKey int

//go:generate stringer -type=ContextKey -output=contextkey.gen.go -linecomment
const (
	ContextKeyRequestID ContextKey = iota + 1 // request_id
	ContextKeyTraceID                         // trace_id
	ContextKeySpanID                          // span_id
)
