package orkle

import "github.com/gnolang/gno/agent2/p/orkle/ingester"

type Ingester interface {
	Type() ingester.Type
	Ingest(value, providerAddress string) (canAutoCommit bool)
	CommitValue(storage Storage, providerAddress string)
}
