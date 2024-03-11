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
		if c.MeterName != "" {
			c.MeterName = meterName
		}
	}
}

func WithOptionExporterEndpoint(exporterEndpoint string) Option {
	return func(c *options.Config) {
		if c.ExporterEndpoint != "" {
			c.ExporterEndpoint = exporterEndpoint
		}
	}
}

func WithOptionFakeMetrics() Option {
	return func(c *options.Config) {
		c.UseFakeMetrics = true
	}
}

func WithOptionServiceName(serviceName string) Option {
	return func(c *options.Config) {
		if c.ServiceName != "" {
			c.ServiceName = serviceName
		}
	}
}
