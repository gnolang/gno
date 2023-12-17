package task

type Aggregator interface {
	// Aggregate runs an aggregation on all values written using the AddValue method.
	// It returns the aggregated value as a string.
	Aggregate() string
	AddValue(value string)
}
