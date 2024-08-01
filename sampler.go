package otelutil

import (
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type AlwaysErrorSampler struct {
	base sdktrace.Sampler
}

func (cs AlwaysErrorSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	// Check if the span has an error status
	for _, attr := range p.Attributes {
		if attr.Key == attribute.Key("status") && attr.Value.AsString() == "error" {
			return sdktrace.SamplingResult{
				Decision:   sdktrace.RecordAndSample,
				Attributes: p.Attributes,
			}
		}
	}
	// Fallback to base/default sampling
	return cs.base.ShouldSample(p)
}

func (cs AlwaysErrorSampler) Description() string {
	return "AlwaysErrorSampler with TraceIDRatioBased and always sample errors"
}

func NewAlwaysErrorSampler(base sdktrace.Sampler) sdktrace.Sampler {
	return AlwaysErrorSampler{
		base: base,
	}
}
