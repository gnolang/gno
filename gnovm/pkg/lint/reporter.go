package lint

type Reporter interface {
	Report(issue Issue)
	Flush() error
	Summary() (info, warnings, errors int)
}

// Factory pattern deferred - add when we have multiple reporters (JSON, SARIF)
// type ReporterFactory func(w io.Writer) Reporter
// var reporterFactories = make(map[string]ReporterFactory)
// func RegisterReporter(format string, factory ReporterFactory) { ... }
// func NewReporter(format string, w io.Writer) (Reporter, error) { ... }
// func AvailableFormats() []string { ... }
