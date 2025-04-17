package telemetry

import (
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type TracerFactory func() trace.Tracer

var noopTracerProvider = noop.NewTracerProvider()

func Tracer(name string, options ...trace.TracerOption) TracerFactory {
	var once sync.Once
	var t trace.Tracer = noopTracerProvider.Tracer(name, options...) // Initilize noop tracer as default
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
