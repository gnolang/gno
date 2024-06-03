// Package config contains the configuration types and helpers for the telemetry
// package.
package config

// Config is the configuration struct for the tm2 telemetry package.
type Config struct {
	MetricsEnabled   bool   `toml:"enabled"`
	MeterName        string `toml:"meter_name"`
	ServiceName      string `toml:"service_name"`
	ExporterEndpoint string `toml:"exporter_endpoint" comment:"the endpoint to export metrics to, like a local OpenTelemetry collector"`
}

// DefaultTelemetryConfig is the default configuration used for the node.
func DefaultTelemetryConfig() *Config {
	return &Config{
		MetricsEnabled:   false,
		MeterName:        "gno.land",
		ServiceName:      "gno.land",
		ExporterEndpoint: "",
	}
}

// TestTelemetryConfig is the test configuration. Currently it is an alias for
// [DefaultTelemetryConfig].
func TestTelemetryConfig() *Config {
	return DefaultTelemetryConfig()
}
