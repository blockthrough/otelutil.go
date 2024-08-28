package otelutil

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Tracer is an alias for oteltrace.Tracer.
// to simplify the import path for the user.
type Tracer = oteltrace.Tracer
type TracerProvider = trace.TracerProvider
type ReadOnlySpan = trace.ReadOnlySpan
type SpanExporter = trace.SpanExporter

// Use SpanFromContext to get the span from the context.
var SpanFromContext = oteltrace.SpanFromContext

var AttrString = attribute.String
var AttrInt64 = attribute.Int64
var AttrInt = attribute.Int

var WithFilter = otelhttp.WithFilter
var WithAttributes = oteltrace.WithAttributes

// Get the Tracer object based on the name
// for example:
//
//	tracer := trace.Get("prebid-server/trace")
//	err = func(ctx context.Context) error {
//		ctx, span := tracer.Start(ctx, "test")
//		defer span.End()
//
//		_ = ctx
//
//		return nil
//	}(ctx)
//	if err != nil {
//		 return err
//	}
func Get(name string) Tracer {
	return otel.GetTracerProvider().Tracer(name)
}

// GetDeferedTracer returns a function that returns the Tracer object based on the name.
// usually this function needs to be called only once at global scope of the package
// and te reason it returns a function is that to defer getting tracer object until
// it has been initliazed. Usually initialization of the tracer object is done in the
// main function and it requires some time to initialize the object globally
func GetDeferedTracer(name string) func() Tracer {
	var tracer Tracer
	var once sync.Once

	return func() Tracer {
		once.Do(func() {
			tracer = Get(name)
		})
		return tracer
	}
}

func RecordError(span oteltrace.Span, err *error) {
	if err != nil && *err != nil {
		span.SetStatus(codes.Error, "error recorded")
		span.RecordError(*err)
	}
}

func Start(ctx context.Context, tracer Tracer, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	return tracer.Start(ctx, name, opts...)
}

// Finish a span and record the error if any, this is a helper function
// to simplify the code.
func Finish(span oteltrace.Span, err *error) {
	RecordError(span, err)
	span.End()
}

var NewTransport = otelhttp.NewTransport

type traceOpt struct {
	name                 string
	sampleRate           float64
	spanProcessorWrapper SpanProcessorWrapper
	exporter             SpanExporter
	spanStartAttributes  []attribute.KeyValue
}

type TraceOption func(*traceOpt)

func WithSampleRate(rate float64) TraceOption {
	return func(opt *traceOpt) {
		opt.sampleRate = rate
	}
}

func WithName(name string) TraceOption {
	return func(opt *traceOpt) {
		opt.name = name
	}
}

func WithSpanProcessor(spw SpanProcessorWrapper) TraceOption {
	return func(opt *traceOpt) {
		opt.spanProcessorWrapper = spw
	}
}

func WithExporter(exporter SpanExporter) TraceOption {
	return func(opt *traceOpt) {
		opt.exporter = exporter
	}
}

func WithSpanStartAttributes(spanStartAttributes ...attribute.KeyValue) TraceOption {
	return func(opt *traceOpt) {
		opt.spanStartAttributes = spanStartAttributes
	}
}

func SetupTraceOTEL(ctx context.Context, optFns ...TraceOption) (*trace.TracerProvider, error) {
	opt := traceOpt{
		name:       "default-name",
		sampleRate: 1.0,
	}

	for _, fn := range optFns {
		fn(&opt)
	}

	if opt.exporter == nil {
		return nil, errors.New("exporter is required")
	}

	// Identify your application using resource detection
	res, err := resource.New(ctx,
		// Use the GCP resource detector to detect information about the GCP platform
		resource.WithDetectors(gcp.NewDetector()),
		// Keep the default detectors
		resource.WithTelemetrySDK(),
		// Add your own custom attributes to identify your application
		resource.WithAttributes(
			semconv.ServiceNameKey.String(opt.name),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// The ParentBased sampler respects the sampling decision made by the parent span,
	// ensuring that once the parent span is sampled, all child spans are also sampled.
	baseSampler := trace.ParentBased(trace.TraceIDRatioBased(opt.sampleRate))

	// The sampler is a decorator that adds additional logic to the base sampler to determine if a span should be sampled
	sampler := NewSampler(baseSampler, opt.spanStartAttributes...)
	providerOpts := []trace.TracerProviderOption{
		trace.WithResource(res),
		trace.WithSampler(sampler),
	}

	if opt.spanProcessorWrapper != nil {
		providerOpts = append(
			providerOpts,
			trace.WithSpanProcessor(
				opt.spanProcessorWrapper(
					trace.NewBatchSpanProcessor(opt.exporter),
				),
			),
		)
	} else {
		providerOpts = append(
			providerOpts,
			trace.WithBatcher(opt.exporter),
		)
	}

	// Configure Metric Export to send metrics as OTLP
	mreader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, err
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(mreader),
	)

	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tp := trace.NewTracerProvider(providerOpts...)

	return tp, nil
}
