// Package tracer provides support for distributed tracing.
package tracer

import (
	"log"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/propagation"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Init creates a new trace provider instance and registers it as global trace provider.
func Init(serviceName string, reporterURI string, probability float64, log *log.Logger) error {
	exporter, err := zipkin.NewRawExporter(
		reporterURI,
		serviceName,
		zipkin.WithLogger(log),
	)
	if err != nil {
		return errors.Wrap(err, "creating new exporter")
	}

	// Demo mode configuarion, always record and sample.
	sampler := sdktrace.AlwaysSample()
	if probability < 1 {
		// Production mode configuarion. A probability=0.05 means only 5% of
		// tracing information will be exported to Zipkin.
		sampler = sdktrace.ProbabilitySampler(.10)
	}
	tp, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sampler}),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
			sdktrace.WithBatchTimeout(sdktrace.DefaultBatchTimeout),
			sdktrace.WithMaxExportBatchSize(sdktrace.DefaultMaxExportBatchSize),
		),
	)
	if err != nil {
		return errors.Wrap(err, "creating new provider")
	}

	global.SetTraceProvider(tp)

	var b3 trace.B3
	props := propagation.New(propagation.WithExtractors(b3))
	global.SetPropagators(props)

	return nil
}
