package gnorkle

import "gno.land/p/demo/gnorkle/v1/ingester"

type Ingester interface {
	Type() ingester.Type
	Ingest(value, providerAddress string) (canAutoCommit bool)
	CommitValue(storage Storage, providerAddress string)
}
