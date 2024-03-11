package options

type Config struct {
	MetricsEnabled   bool
	UseFakeMetrics   bool
	MeterName        string
	ServiceName      string
	ExporterEndpoint string
}
