package multi

import (
	"strings"

	"gno.land/p/demo/gnorkle/gnorkle"
	"gno.land/p/demo/gnorkle/ingester"
)

// ValueIngester is an ingester that ingests multiple values.
type ValueIngester struct {
	values []string
}

// Type returns the type of the ingester.
func (i *ValueIngester) Type() ingester.Type {
	return ingester.TypeMulti
}

// Ingest ingests a value provided by the given agent address.
func (i *ValueIngester) Ingest(value, providerAddress string) (bool, error) {
	if i == nil {
		return false, ingester.ErrUndefined
	}

	i.values = append(i.values, value)
	return false, nil
}

// CommitValue commits the ingested values to the given storage instance.
func (i *ValueIngester) CommitValue(valueStorer gnorkle.Storage, providerAddress string) error {
	if i == nil {
		return ingester.ErrUndefined
	}

	// append all values
	return valueStorer.Put(strings.Join(i.values, ","))
}
