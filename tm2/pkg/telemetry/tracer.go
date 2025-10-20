package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var noopTracerProvider = noop.NewTracerProvider()

func Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	if !TracesEnabled() {
		// noop provides an implementation of the OpenTelemetry trace API
		// that produces no telemetry and minimizes used computation resources.
		return noopTracerProvider.Tracer(name, options...)

	}

	provider := otel.GetTracerProvider()
	return provider.Tracer(name, options...)
}
