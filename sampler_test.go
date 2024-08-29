package otelutil_test

import (
	"context"
	"testing"

	"github.com/blockthrough/otelutil.go"
	"github.com/stretchr/testify/assert"
)

func TestSamplerWithSpanStartFilter(t *testing.T) {
	t.Parallel()

	keyVal := otelutil.AttrString("test-attr", "test-val")

	tp, getSpans := createTestTraceProvider(t, nil, keyVal)

	tracer := tp.Tracer("test-tracer")

	ctx := context.Background()

	// Start a span with a name and attribute name matching the attribute we want to filter
	ctx, span := tracer.Start(ctx, "test-span-1", otelutil.WithAttributes(keyVal))
	span.End()

	// Start a span with a name and attribute name NOT matching the attribute we want to filter
	ctx, span = tracer.Start(ctx, "test-span-2", otelutil.WithAttributes(otelutil.AttrString("test-attr-2", "test-val-2")))
	span.End()

	tp.ForceFlush(ctx)

	// Fetch the spans from the TracerProvider
	spans := getSpans()

	assert.Len(t, spans, 1)

	assert.Equal(t, "test-span-1", spans[0].Name)
}

func TestSamplerWithDefaultSpanStartFilter(t *testing.T) {
	t.Parallel()

	keyVal := otelutil.AttrString("test-attr", "test-val")

	tp, getSpans := createTestTraceProvider(t, nil, keyVal)

	otelutil.SetWithDefaultSpanStartAttributes(keyVal)

	tracer := tp.Tracer("test-tracer")

	ctx := context.Background()

	// Start a span with a name and default attribute name matching the attribute we want to filter
	ctx, span := tracer.Start(ctx, "test-span-1", otelutil.GetWithDefaultSpanStartAttributes())
	span.End()

	// Start a span with a name and attribute name NOT matching the attribute we want to filter
	ctx, span = tracer.Start(ctx, "test-span-2", otelutil.WithAttributes(otelutil.AttrString("test-attr-2", "test-val-2")))
	span.End()

	tp.ForceFlush(ctx)

	// Fetch the spans from the TracerProvider
	spans := getSpans()

	assert.Len(t, spans, 1)

	assert.Equal(t, "test-span-1", spans[0].Name)
}
