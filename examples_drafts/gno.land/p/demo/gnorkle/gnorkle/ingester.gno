package gnorkle

import "gno.land/p/demo/gnorkle/ingester"

// Ingester is the abstraction that allows a `Feed` to ingest data from agents
// and commit it to storage using zero or more intermediate aggregation steps.
type Ingester interface {
	Type() ingester.Type
	Ingest(value, providerAddress string) (canAutoCommit bool, err error)
	CommitValue(storage Storage, providerAddress string) error
}
