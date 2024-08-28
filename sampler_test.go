package otelutil_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestSamplerWithSpanStartFilter(t *testing.T) {
	t.Parallel()

	keyVal := attribute.KeyValue{
		Key:   "test-attr",
		Value: attribute.StringValue("test-val"),
	}
	tp, getSpans := createTestTraceProvider(t, nil, keyVal)

	tracer := tp.Tracer("test-tracer")

	ctx := context.Background()

	// Start a span with a name and attribute name matching the attribute we want to filter
	ctx, span := tracer.Start(ctx, "test-span-1", trace.WithAttributes(keyVal))
	span.End()

	// Start a span with a name and attribute name NOT matching the attribute we want to filter
	ctx, span = tracer.Start(ctx, "test-span-2", trace.WithAttributes(attribute.KeyValue{
		Key:   "test-attr-2",
		Value: attribute.StringValue("test-val-2"),
	}))
	span.End()

	tp.ForceFlush(ctx)

	// Fetch the spans from the TracerProvider
	spans := getSpans()

	assert.Len(t, spans, 1)

	assert.Equal(t, "test-span-1", spans[0].Name)
}
