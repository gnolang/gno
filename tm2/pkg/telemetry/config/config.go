package config

import (
	"errors"
	"os"
)

var errEndpointNotSet = errors.New("telemetry exporter endpoint not set")

// Config is the configuration struct for the tm2 telemetry package
type Config struct {
	MetricsEnabled   bool   `toml:"enabled"`
	MeterName        string `toml:"meter_name"`
	ServiceName      string `toml:"service_name"`
	ServiceInstance  string `toml:"service_instance"`
	ExporterEndpoint string `toml:"exporter_endpoint" comment:"the endpoint to export metrics to, like a local OpenTelemetry collector"`
}

// DefaultTelemetryConfig is the default configuration used for the node
func DefaultTelemetryConfig() *Config {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "gno-node"
	}
	return &Config{
		MetricsEnabled:   false,
		MeterName:        "gno.land",
		ServiceName:      "gno.land",
		ServiceInstance:  hostname,
		ExporterEndpoint: "",
	}
}

// ValidateBasic performs basic telemetry config validation and
// returns an error if any check fails
func (cfg *Config) ValidateBasic() error {
	if cfg.ExporterEndpoint == "" {
		return errEndpointNotSet
	}

	return nil
}
