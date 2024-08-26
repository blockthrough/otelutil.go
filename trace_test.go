package otelutil_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/blockthrough/otelutil.go"
)

func createTestTraceProvider(t *testing.T) (*otelutil.TracerProvider, func() tracetest.SpanStubs) {
	// Create a TracerProvider using the Tracetest SDK
	exp := tracetest.NewInMemoryExporter()
	tp, shutdown, err := otelutil.SetupTraceOTEL(context.Background(), otelutil.WithExporter(exp))
	assert.NoError(t, err)
	t.Cleanup(func() {
		shutdown(context.Background())
	})

	return tp, exp.GetSpans
}

func TestSpanFilter(t *testing.T) {
	tp, getSpans := createTestTraceProvider(t)

	tracer := tp.Tracer("test-tracer")

	// Start a span
	_, span := tracer.Start(context.Background(), "test-span")
	span.SetAttributes(attribute.String("key", "value"))
	span.End()

	tp.ForceFlush(context.Background())

	// Fetch the spans from the TracerProvider
	spans := getSpans()

	assert.Len(t, spans, 1)

	assert.Equal(t, "test-span", spans[0].Name)
}
