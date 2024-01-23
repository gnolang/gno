package single

import (
	"gno.land/p/demo/orkle/v1/ingester"
	"gno.land/p/demo/orkle/v1/orkle"
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
