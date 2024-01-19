package single

import (
	"github.com/gnolang/gno/agent2/p/orkle"
	"github.com/gnolang/gno/agent2/p/orkle/ingester"
)

type ValueIngester struct {
	value string
}

func (i *ValueIngester) Type() ingester.Type {
	return ingester.TypeSingle
}

func (i *ValueIngester) Ingest(value, providerAddress string) bool {
	i.value = value
	return true
}

func (i *ValueIngester) CommitValue(valueStorer orkle.Storage, providerAddress string) {
	valueStorer.Put(i.value)
}
