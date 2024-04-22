package telemetry

import "github.com/gnolang/gno/telemetry/options"

type Option func(*options.Config)

func WithOptionMetricsEnabled() Option {
	return func(c *options.Config) {
		c.MetricsEnabled = true
	}
}

func WithOptionMeterName(meterName string) Option {
	return func(c *options.Config) {
		if meterName != "" {
			c.MeterName = meterName
		}
	}
}

func WithOptionExporterEndpoint(exporterEndpoint string) Option {
	return func(c *options.Config) {
		if exporterEndpoint != "" {
			c.ExporterEndpoint = exporterEndpoint
		}
	}
}

func WithOptionServiceName(serviceName string) Option {
	return func(c *options.Config) {
		if serviceName != "" {
			c.ServiceName = serviceName
		}
	}
}
