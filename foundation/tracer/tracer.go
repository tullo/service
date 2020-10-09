// Package tracer provides support for distributed tracing.
package tracer

import (
	"log"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	"go.opentelemetry.io/otel/sdk/trace"
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
	sampler := trace.AlwaysSample()
	if probability < 1 {
		// Production mode configuarion. A probability=0.05 means only 5% of
		// tracing information will be exported to Zipkin.
		sampler = trace.TraceIDRatioBased(.10)
	}
	tp := trace.NewTracerProvider(
		trace.WithConfig(trace.Config{DefaultSampler: sampler}),
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultBatchTimeout),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
	)

	global.SetTracerProvider(tp)

	//	var b3 trace.B3
	//	props := propagation.New(propagation.WithExtractors(b3))
	//	global.SetPropagators(props)

	return nil
}
