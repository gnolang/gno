package single

import (
	"gno.land/p/demo/gnorkle/v1/gnorkle"
	"gno.land/p/demo/gnorkle/v1/ingester"
)

// ValueIngester is an ingester that ingests a single value.
type ValueIngester struct {
	value string
}

// Type returns the type of the ingester.
func (i *ValueIngester) Type() ingester.Type {
	return ingester.TypeSingle
}

// Ingest ingests a value provided by the given agent address.
func (i *ValueIngester) Ingest(value, providerAddress string) bool {
	i.value = value
	return true
}

// CommitValue commits the ingested value to the given storage instance.
func (i *ValueIngester) CommitValue(valueStorer gnorkle.Storage, providerAddress string) {
	valueStorer.Put(i.value)
}
