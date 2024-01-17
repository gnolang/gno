package single

import (
	"github.com/gnolang/gno/agent2/p/orkle"
	"github.com/gnolang/gno/agent2/p/orkle/ingester"
)

type SingleValueIngester struct {
	value string
}

func (i *SingleValueIngester) Type() ingester.Type {
	return ingester.TypeSingle
}

func (i *SingleValueIngester) Ingest(value, providerAddress string) {
	i.value = value
}

func (i *SingleValueIngester) CommitValue(valueStorer orkle.Storage, providerAddress string) {
	valueStorer.Put(i.value)
}
