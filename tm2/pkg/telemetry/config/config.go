package config

import (
	"errors"
)

var errEndpointNotSet = errors.New("telemetry exporter endpoint not set")

// Config is the configuration struct for the tm2 telemetry package
type Config struct {
	MetricsEnabled    bool   `json:"metrics_enabled" toml:"metrics_enabled"`
	MeterName         string `json:"meter_name" toml:"meter_name"`
	ServiceName       string `json:"service_name" toml:"service_name" comment:"in Prometheus this is transformed into the label 'exported_job'"`
	ServiceInstanceID string `json:"service_instance_id" toml:"service_instance_id" comment:"the ID helps to distinguish instances of the same service that exist at the same time (e.g. instances of a horizontally scaled service), in Prometheus this is transformed into the label 'exported_instance"`
	ExporterEndpoint  string `json:"exporter_endpoint" toml:"exporter_endpoint" comment:"the endpoint to export metrics to, like a local OpenTelemetry collector"`
	TracesEnabled     bool   `json:"traces_enabled" toml:"traces_enabled"`
}

// DefaultTelemetryConfig is the default configuration used for the node
func DefaultTelemetryConfig() *Config {
	return &Config{
		MetricsEnabled:    false,
		MeterName:         "tm2",
		ServiceName:       "tm2",
		ServiceInstanceID: "tm2-node-1",
		ExporterEndpoint:  "",
		TracesEnabled:     false,
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
