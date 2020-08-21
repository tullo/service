// Package tracer provides support for distributed tracing.
package tracer

import (
	"log"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
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

	// WARNING: The current Init settings are using defaults which I listed out
	// for readability. Please review the documentation for opentelemetry.
	tp, err := sdktrace.NewProvider(
		// Demo mode configuarion, always sample.
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		// Production mode configuarion
		// sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.ProbabilitySampler(.10)}),
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
	return nil
}
