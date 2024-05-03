package single

import (
	"gno.land/p/demo/gnorkle/gnorkle"
	"gno.land/p/demo/gnorkle/ingester"
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
func (i *ValueIngester) Ingest(value, providerAddress string) (bool, error) {
	if i == nil {
		return false, ingester.ErrUndefined
	}

	i.value = value
	return true, nil
}

// CommitValue commits the ingested value to the given storage instance.
func (i *ValueIngester) CommitValue(valueStorer gnorkle.Storage, providerAddress string) error {
	if i == nil {
		return ingester.ErrUndefined
	}

	return valueStorer.Put(i.value)
}
