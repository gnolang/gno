# Telemetry

The purpose of this package is to provide a way to easily integrate OpenTelemetry Protocol (OTLP) metrics collection into our codebase.

Metrics can be enabled using environment variables. The following variables are supported:
- `TELEM_METRICS_ENABLED`: setting to `true` will enable metrics collection
- `TELEM_USE_FAKE_METRICS`; optional; setting to `true` will collect dummy values. This can be good for testing that collection is working.
- `TELEM_METER_NAME`: optionally set the meter name; the default is `gno.land`
- `TELEM_SERVICE_NAME`: optionally set the service name; the default is `gno.land`
- `TELEM_EXPORTER_ENDPOINT`: required; this is the endpoint to export metrics to