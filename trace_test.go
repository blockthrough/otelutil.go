package otelutil_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/blockthrough/otelutil.go"
)

func TestSpanFilter(t *testing.T) {
	// Create a TracerProvider using the Tracetest SDK
	exp := tracetest.NewInMemoryExporter()

	tp, shutdown, err := otelutil.SetupTraceOTEL(context.Background(), otelutil.WithExporter(exp))
	assert.NoError(t, err)
	t.Cleanup(func() {
		shutdown(context.Background())
	})

	tracer := tp.Tracer("test-tracer")

	// Start a span
	_, span := tracer.Start(context.Background(), "test-span")
	span.SetAttributes(attribute.String("key", "value"))
	span.End()

	tp.ForceFlush(context.Background())

	// Fetch the spans from the TracerProvider
	spans := exp.GetSpans()

	// Apply your span filter here and assert the results
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	// Further checks on the span
	if spans[0].Name != "test-span" {
		t.Errorf("expected span name to be 'test-span', got %s", spans[0].Name)
	}
}
