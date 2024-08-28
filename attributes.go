package otelutil

import (
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var WithAttributes = oteltrace.WithAttributes

var AttrString = attribute.String
var AttrInt64 = attribute.Int64
var AttrInt = attribute.Int

var DefaultSpanStartAttributes = []attribute.KeyValue{}

func SetWithDefaultSpanStartAttributes(attrs ...attribute.KeyValue) {
	DefaultSpanStartAttributes = attrs
}

func GetWithDefaultSpanStartAttributes() oteltrace.SpanStartEventOption {
	return WithAttributes(DefaultSpanStartAttributes...)
}
