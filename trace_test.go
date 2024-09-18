package otelutil_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/blockthrough/otelutil.go"
)

func createTestTraceProvider(t *testing.T, filterFn otelutil.SpanProcessorWrapper, spanStartAttributes ...attribute.KeyValue) (*otelutil.TracerProvider, func() tracetest.SpanStubs) {
	// Create a TracerProvider using the Tracetest SDK
	ctx := context.Background()
	exp := tracetest.NewInMemoryExporter()
	tp, shutdown, err := otelutil.SetupTraceOTEL(
		ctx,
		otelutil.WithExporter(exp),
		otelutil.WithSpanProcessor(filterFn),
		otelutil.WithDefaultSpanStartAttributes(spanStartAttributes...),
		otelutil.WithNotSetDefaultTracer(),
	)
	assert.NoError(t, err)
	t.Cleanup(func() {
		shutdown(ctx)
	})

	return tp, exp.GetSpans
}

func TestSpanFilter(t *testing.T) {
	t.Parallel()

	tp, getSpans := createTestTraceProvider(t, nil)

	tracer := tp.Tracer("test-tracer")

	ctx := context.Background()

	// Start a span with a name
	ctx, span := tracer.Start(ctx, "test-span-1")
	span.End()

	ctx, span = tracer.Start(ctx, "test-span-2")
	span.End()

	tp.ForceFlush(ctx)

	// Fetch the spans from the TracerProvider
	spans := getSpans()

	assert.Len(t, spans, 2)

	assert.Equal(t, "test-span-1", spans[0].Name)
	assert.Equal(t, "test-span-2", spans[1].Name)
}

func getAttribute(attributes []attribute.KeyValue, key string) (attribute.Value, bool) {
	for _, a := range attributes {
		if string(a.Key) == key {
			return a.Value, true
		}
	}

	return attribute.Value{}, false
}

func TestError(t *testing.T) {
	t.Parallel()

	tp, getSpans := createTestTraceProvider(t, nil)

	tracer := tp.Tracer("test-tracer")

	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span-1")
	err := fmt.Errorf("test-error")

	otelutil.Finish(span, &err)

	tp.ForceFlush(ctx)

	spans := getSpans()

	assert.Len(t, spans, 1)

	s := spans[0]

	hasError, ok := getAttribute(s.Attributes, "has_error")
	assert.True(t, ok)
	assert.Equal(t, true, hasError.AsBool())

	errMsg, ok := getAttribute(s.Attributes, "error_message")
	assert.True(t, ok)
	assert.Equal(t, "test-error", errMsg.AsString())
}
