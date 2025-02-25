package tracing

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gnolang/gno/tm2/pkg/telemetry/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

func Init(cfg config.Config) error {
	var (
		ctx = context.Background()
		exp sdkTrace.SpanExporter
	)

	u, err := url.Parse(cfg.ExporterEndpoint)
	if err != nil {
		return fmt.Errorf("error parsing tracer exporter endpoint: %s, %w", cfg.ExporterEndpoint, err)
	}

	// Use oltp metric exporter with http/https or grpc
	switch u.Scheme {
	case "http", "https":
		exp, err = otlptracehttp.New(
			ctx,
			otlptracehttp.WithEndpointURL(cfg.ExporterEndpoint),
		)
		if err != nil {
			return fmt.Errorf("unable to create http tracing exporter, %w", err)
		}
	default:
		exp, err = otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithEndpoint(cfg.ExporterEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		if err != nil {
			return fmt.Errorf("unable to create grpc tracing exporter, %w", err)
		}
	}

	traceProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(exp),
		sdkTrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("tm2"),
		)),
	)

	otel.SetTracerProvider(traceProvider)

	return nil
}
