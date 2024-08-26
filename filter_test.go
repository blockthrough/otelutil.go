package otelutil_test

import (
	"context"
	"testing"

	"github.com/blockthrough/otelutil.go"
	"github.com/stretchr/testify/assert"
)

func TestSpanFilterOnlyNames(t *testing.T) {
	tp, getSpans := createTestTraceProvider(t, otelutil.WithSpanFilterOnlyNames("test-span-1"))

	tracer := tp.Tracer("test-tracer")

	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span-1")
	span.End()

	ctx, span = tracer.Start(ctx, "test-span-2")
	span.End()

	tp.ForceFlush(ctx)

	spans := getSpans()

	assert.Len(t, spans, 1)
	assert.Equal(t, "test-span-1", spans[0].Name)
}

func TestSpanFilterOnlyScopeNames(t *testing.T) {
	tp, getSpans := createTestTraceProvider(t, otelutil.WithSpanFilterOnlyScopeNames("test-tracer-1"))

	{
		tracer := tp.Tracer("test-tracer-1")

		ctx := context.Background()

		_, span := tracer.Start(ctx, "test-span-1")
		span.End()
	}

	{
		tracer := tp.Tracer("test-tracer-2")

		ctx := context.Background()

		_, span := tracer.Start(ctx, "test-span-2")
		span.End()
	}

	tp.ForceFlush(context.Background())

	spans := getSpans()

	assert.Len(t, spans, 1)
	assert.Equal(t, "test-span-1", spans[0].Name)
}

func TestMultipleSpanFilters(t *testing.T) {
	tp, getSpans := createTestTraceProvider(
		t,
		otelutil.WithMultipleSpanFilters(
			otelutil.WithSpanFilterOnlyScopeNames("test-tracer-1"),
			otelutil.WithSpanFilterOnlyNames("test-span-1", "test-span-2"),
		),
	)

	tracer := tp.Tracer("test-tracer-1")

	ctx := context.Background()

	ctx, span := tracer.Start(ctx, "test-span-1")
	span.End()

	ctx, span = tracer.Start(ctx, "test-span-2")
	span.End()

	ctx, span = tracer.Start(ctx, "test-span-3")
	span.End()

	tp.ForceFlush(ctx)

	spans := getSpans()

	assert.Len(t, spans, 2)
	assert.Equal(t, "test-span-1", spans[0].Name)
	assert.Equal(t, "test-span-2", spans[1].Name)
}
