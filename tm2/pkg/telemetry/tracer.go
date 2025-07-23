package telemetry

import (
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

var noopTracerProvider = noop.NewTracerProvider()

func Tracer(name string, options ...trace.TracerOption) func() trace.Tracer {
	var once sync.Once
	t := noopTracerProvider.Tracer(name, options...) // Initilize noop tracer as default
	return func() trace.Tracer {
		if TracesEnabled() {
			once.Do(func() {
				provider := otel.GetTracerProvider()
				t = provider.Tracer(name, options...)
			})
		}

		return t
	}
}
