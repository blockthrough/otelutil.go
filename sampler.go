package otelutil

import (
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// sampler -  if the span has the specified attributes and the base sampler would have sampled it, then it will be sampled.
type sampler struct {
	base  sdktrace.Sampler
	attrs map[attribute.Key]attribute.Value
}

func (cs sampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	result := cs.base.ShouldSample(p)
	if result.Decision != sdktrace.RecordAndSample {
		return result
	}

	if len(cs.attrs) == 0 {
		return result
	}

	psc := trace.SpanContextFromContext(p.ParentContext)

	for _, attr := range p.Attributes {
		if val, ok := cs.attrs[attr.Key]; ok {
			if val.AsString() == attr.Value.AsString() {
				return sdktrace.SamplingResult{
					Decision:   sdktrace.RecordAndSample,
					Attributes: p.Attributes,
					Tracestate: psc.TraceState(),
				}
			}
		}
	}

	return sdktrace.SamplingResult{Decision: sdktrace.Drop, Attributes: p.Attributes, Tracestate: psc.TraceState()}
}

func (cs sampler) Description() string {
	return "sampler with base sampler: " + cs.base.Description()
}

func NewSampler(base sdktrace.Sampler, attrs ...attribute.KeyValue) sdktrace.Sampler {
	m := map[attribute.Key]attribute.Value{}
	for _, attr := range attrs {
		m[attr.Key] = attr.Value
	}
	return sampler{
		base:  base,
		attrs: m,
	}
}
