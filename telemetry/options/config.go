package options

type Config struct {
	MetricsEnabled   bool
	TracesEnabled    bool
	UseFakeMetrics   bool
	Port             uint64
	MeterName        string
	ServiceName      string
	ExporterEndpoint string
	TraceType        int64
}
