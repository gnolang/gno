# Telemetry

The purpose of this package is to provide a way to easily integrate OpenTelemetry Protocol (OTLP) metrics collection into our codebase.

## Configure environment variables
Metrics can be enabled using environment variables. The following variables are supported:
- `TELEM_METRICS_ENABLED`: setting to `true` will enable metrics collection
- `TELEM_METER_NAME`: optionally set the meter name; the default is `gno.land`
- `TELEM_SERVICE_NAME`: optionally set the service name; the default is `gno.land`
- `TELEM_EXPORTER_ENDPOINT`: required; this is the endpoint to export metrics to, like a local OTEL collector

## OTEL configuration
There are many ways configure the OTEL pipeline for exporting metrics. Here is an example of how a local OTEL collector can be configured to send metrics to Grafana Cloud. This is an optional step and can be highly customized.

### OTEL collector
The latest collector releases can be found [here](https://github.com/open-telemetry/opentelemetry-collector-releases/releases). This is an example of the config that can be used to receive metrics from gno.land and publish them to Grafana Cloud.
```yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317 # should be the same as the TELEM_EXPORTER_ENDPOINT variable

processors:
  batch:

exporters:
  otlphttp:
    endpoint: https://otlp-gateway-prod-us-east-0.grafana.net/otlp

service:
  pipelines:
    metrics:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlphttp]
```

Collector exporter environment variables, including those for authentication, can be found [here](https://opentelemetry.io/docs/specs/otel/protocol/exporter/).

## Resources
- https://opentelemetry.io/docs/collector/
- https://grafana.com/docs/grafana-cloud/monitor-applications/application-observability/setup/collector/