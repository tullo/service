// Package tracer provides support for distributed tracing.
package tracer

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// Config holds trace config properties.
type Config struct {
	ServiceName string
	ReporterURI string
	Probability float64
}

// Init creates a new trace provider instance and registers it as global trace provider.
func Init(l *log.Logger, c *Config) (func(ctx context.Context) error, error) {
	exporter, err := zipkin.New(c.ReporterURI, zipkin.WithLogger(l))
	if err != nil {
		return nil, errors.Wrap(err, "creating new exporter")
	}

	batcher := sdktrace.NewBatchSpanProcessor(
		exporter,
		sdktrace.WithBatchTimeout(sdktrace.DefaultScheduleDelay*time.Millisecond),
		sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
	)

	// Demo mode configuarion, always record and sample.
	//
	// By default the returned TracerProvider is configured with:
	// - a ParentBased(AlwaysSample) Sampler
	// - a random number IDGenerator
	// - the resource.Default() Resource
	// - the default SpanLimits
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(batcher),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(c.ServiceName),
		)),
	)

	if c.Probability < 1 {
		// Production mode configuarion. A probability=0.01 means only 1% of
		// tracing information will be exported to Zipkin.
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithSpanProcessor(batcher),
			sdktrace.WithSampler(sdktrace.TraceIDRatioBased(c.Probability)),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("zipkin-search"),
			)))
	}

	otel.SetTracerProvider(tp)

	//	var b3 trace.B3
	//	props := propagation.New(propagation.WithExtractors(b3))
	//	global.SetPropagators(props)

	return tp.Shutdown, nil
}
