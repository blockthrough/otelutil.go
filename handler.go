package otelutil

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func WithSpanStartAttrs(attrs ...attribute.KeyValue) otelhttp.Option {
	return otelhttp.WithSpanOptions(trace.WithAttributes(attrs...))
}

func NewHandler(handler http.Handler, operation string, opts ...otelhttp.Option) http.Handler {
	return otelhttp.NewMiddleware(operation, opts...)(handler)
}
