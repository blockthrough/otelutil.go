package otelutil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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
		span.RecordError(*err)
	}
}

// Finish a span and record the error if any, this is a helper function
// to simplify the code.
func Finish(span oteltrace.Span, err *error) {
	RecordError(span, err)
	span.End()
}

func NewHandler(handler http.Handler, operation string, opts ...otelhttp.Option) http.Handler {
	return otelhttp.NewMiddleware(operation, opts...)(handler)
}

var NewTransport = otelhttp.NewTransport

type traceOpt struct {
	name                 string
	sampleRate           float64
	spanProcessorWrapper SpanProcessorWrapper
	exporter             SpanExporter
	setDefaultTracer     bool
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

func WithNotSetDefaultTracer() TraceOption {
	return func(opt *traceOpt) {
		opt.setDefaultTracer = false
	}
}

func SetupTraceOTEL(ctx context.Context, optFns ...TraceOption) (tp *trace.TracerProvider, shutdown func(context.Context) error, err error) {
	opt := traceOpt{
		name:             "default-name",
		sampleRate:       1.0,
		setDefaultTracer: true,
	}

	for _, fn := range optFns {
		fn(&opt)
	}

	if opt.exporter == nil {
		return nil, nil, errors.New("exporter is required")
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
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var shutdownFuncs []func(context.Context) error

	// shutdown combines shutdown functions from multiple OpenTelemetry
	// components into a single function.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	providerOpts := []trace.TracerProviderOption{
		trace.WithSampler(
			// NewAlwaysErrorSampler is a custom sampler that samples all error spans
			// and delegates the sampling decision for non-error spans to the base sampler.
			NewAlwaysErrorSampler(
				// The ParentBased sampler respects the sampling decision made by the parent span,
				// ensuring that once a trace is sampled, all spans within that trace are also sampled.
				trace.ParentBased(
					trace.TraceIDRatioBased(opt.sampleRate),
				),
			),
		),
		trace.WithResource(res),
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

	tp = trace.NewTracerProvider(providerOpts...)
	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)

	// This option usually needs to be set to false, if we need t test the code
	// in production environment, we need to set this to true using WithDefaultTracer
	if opt.setDefaultTracer {
		otel.SetTracerProvider(tp)
	}

	// Configure Metric Export to send metrics as OTLP
	mreader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		err = errors.Join(err, shutdown(ctx))
		return
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(mreader),
	)
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, shutdown, nil
}
