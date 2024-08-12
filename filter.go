package otelutil

import (
	"context"

	"go.opentelemetry.io/otel/sdk/trace"
)

type filteringSpanProcessor struct {
	namesMap map[string]struct{}
	sp       trace.SpanProcessor
}

var _ trace.SpanProcessor = (*filteringSpanProcessor)(nil)

func (f *filteringSpanProcessor) needsToBeFiltered(spanName string) bool {
	if len(f.namesMap) == 0 {
		return false
	}
	_, ok := f.namesMap[spanName]
	return !ok
}

func (f *filteringSpanProcessor) OnStart(parent context.Context, s trace.ReadWriteSpan) {
	spanName := s.Name()
	if f.needsToBeFiltered(spanName) {
		// No further processing if the span is not an application span
		return
	}

	// Continue processing if it's an application span
	f.sp.OnStart(parent, s)
}

func (f *filteringSpanProcessor) OnEnd(s trace.ReadOnlySpan) {
	spanName := s.Name()
	if f.needsToBeFiltered(spanName) {
		// No further processing if the span is not an application span
		return
	}

	// Continue processing if it's an application span
	f.sp.OnEnd(s)
}

func (f *filteringSpanProcessor) Shutdown(ctx context.Context) error {
	return f.sp.Shutdown(ctx)
}

func (f *filteringSpanProcessor) ForceFlush(ctx context.Context) error {
	return f.sp.ForceFlush(ctx)
}

func WithSpanFilter(sp trace.SpanProcessor, onlyNames ...string) trace.SpanProcessor {
	namesMap := make(map[string]struct{}, len(onlyNames))
	for _, name := range onlyNames {
		namesMap[name] = struct{}{}
	}
	return &filteringSpanProcessor{
		sp:       sp,
		namesMap: namesMap,
	}
}
