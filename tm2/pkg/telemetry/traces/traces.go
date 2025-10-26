package traces

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var provider *sdkTrace.TracerProvider
var tracer trace.Tracer = noop.NewTracerProvider().Tracer("")

func Init(cfg config.Config) error {
	if !cfg.TracesEnabled {
		return nil
	}
	var (
		ctx = context.Background()
		exp sdkTrace.SpanExporter
	)

	u, err := url.Parse(cfg.ExporterEndpoint)
	if err != nil {
		return fmt.Errorf("error parsing tracer exporter endpoint: %s, %w", cfg.ExporterEndpoint, err)
	}

	switch u.Scheme {
	case "http", "https":
		exp, err = otlptracehttp.New(
			ctx,
			otlptracehttp.WithEndpointURL(u.Host),
		)
		if err != nil {
			return fmt.Errorf("unable to create http traces exporter, %w", err)
		}
	case "grpc":
		exp, err = otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(u.Host),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return fmt.Errorf("unable to create grpc metrics exporter, %w", err)
		}
	default:
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	provider = sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(exp),
		sdkTrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String("1.0.0"),
			semconv.ServiceInstanceIDKey.String(cfg.ServiceInstanceID),
		)),
	)

	otel.SetTracerProvider(provider)
	tracer = otel.Tracer(cfg.ServiceName)

	return nil
}

func Shutdown() {
	if provider != nil {
		provider.Shutdown(context.Background())
	}
}

func Tracer() trace.Tracer {
	return tracer
}
