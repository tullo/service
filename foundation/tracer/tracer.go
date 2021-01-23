// Package tracer provides support for distributed tracing.
package tracer

import (
	"log"

	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/zipkin"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Config holds trace config properties.
type Config struct {
	ServiceName string
	ReporterURI string
	Probability float64
}

// Init creates a new trace provider instance and registers it as global trace provider.
func Init(l *log.Logger, c *Config) error {
	exporter, err := zipkin.NewRawExporter(
		c.ReporterURI,
		c.ServiceName,
		zipkin.WithLogger(l),
	)
	if err != nil {
		return errors.Wrap(err, "creating new exporter")
	}

	// Demo mode configuarion, always record and sample.
	sampler := trace.AlwaysSample()
	if c.Probability < 1 {
		// Production mode configuarion. A probability=0.01 means only 1% of
		// tracing information will be exported to Zipkin.
		sampler = trace.TraceIDRatioBased(c.Probability)
	}
	tp := trace.NewTracerProvider(
		trace.WithConfig(trace.Config{DefaultSampler: sampler}),
		trace.WithBatcher(exporter,
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
			trace.WithBatchTimeout(trace.DefaultBatchTimeout),
			trace.WithMaxExportBatchSize(trace.DefaultMaxExportBatchSize),
		),
	)

	otel.SetTracerProvider(tp)

	//	var b3 trace.B3
	//	props := propagation.New(propagation.WithExtractors(b3))
	//	global.SetPropagators(props)

	return nil
}
