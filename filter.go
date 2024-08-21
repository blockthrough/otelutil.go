package otelutil

import (
	"context"

	"go.opentelemetry.io/otel/sdk/trace"
)

// if this function returns false, the span will be filtered out
type SpanFilterFunc func(span trace.ReadOnlySpan) bool

type filteringSpanProcessor struct {
	filterFn SpanFilterFunc
	sp       trace.SpanProcessor
}

var _ trace.SpanProcessor = (*filteringSpanProcessor)(nil)

func (f *filteringSpanProcessor) OnStart(parent context.Context, s trace.ReadWriteSpan) {
	if !f.filterFn(s) {
		return
	}
	f.sp.OnStart(parent, s)
}

func (f *filteringSpanProcessor) OnEnd(s trace.ReadOnlySpan) {
	if !f.filterFn(s) {
		return
	}
	f.sp.OnEnd(s)
}

func (f *filteringSpanProcessor) Shutdown(ctx context.Context) error {
	return f.sp.Shutdown(ctx)
}

func (f *filteringSpanProcessor) ForceFlush(ctx context.Context) error {
	return f.sp.ForceFlush(ctx)
}

func WithSpanFilter(sp trace.SpanProcessor, filterFn SpanFilterFunc) trace.SpanProcessor {
	return &filteringSpanProcessor{
		sp:       sp,
		filterFn: filterFn,
	}
}

func WithSpanFilterOnlyNames(sp trace.SpanProcessor, names ...string) trace.SpanProcessor {
	nameSet := make(map[string]struct{}, len(names))
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	return WithSpanFilter(sp, func(span trace.ReadOnlySpan) bool {
		_, ok := nameSet[span.Name()]
		return ok
	})
}
