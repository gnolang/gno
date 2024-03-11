package traces

import (
	"context"
	"log"

	"github.com/gnolang/gno/telemetry/exporter"
	"github.com/gnolang/gno/telemetry/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc/credentials"

	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type traceFilter int

const (
	traceFilterNone traceFilter = iota
	traceFilterOp
	traceFilterStore
)

var globalTraceFilter traceFilter

func Init(config options.Config) error {

	if config.ExporterEndpoint == "" {
		return exporter.ErrEndpointNotSet
	}

	// TODO: support secure
	var secure bool
	secureOption := otlptracegrpc.WithInsecure()

	if secure {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(config.ExporterEndpoint),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", config.ServiceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		return err
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resources),
		),
	)

	return nil
}

func IsTraceOp() bool {
	return globalTraceFilter == traceFilterNone || globalTraceFilter == traceFilterOp
}

func IsTraceStore() bool {
	return globalTraceFilter == traceFilterNone || globalTraceFilter == traceFilterOp
}
