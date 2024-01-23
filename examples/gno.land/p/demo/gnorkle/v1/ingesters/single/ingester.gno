package single

import (
	"gno.land/p/demo/gnorkle/v1/gnorkle"
	"gno.land/p/demo/gnorkle/v1/ingester"
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

func (i *ValueIngester) CommitValue(valueStorer gnorkle.Storage, providerAddress string) {
	valueStorer.Put(i.value)
}
