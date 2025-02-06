package datastore

import (
	"gno.land/p/moul/collection"
)

const DefaultIndexOptions = collection.DefaultIndex | collection.SparseIndex

type (
	IndexFn func(Record) string

	IndexMultiValueFn func(Record) []string

	Index struct {
		name    string
		options collection.IndexOption
		fn      interface{}
	}
)

func NewIndex(name string, fn IndexFn) Index {
	return Index{
		name:    name,
		options: DefaultIndexOptions,
		fn: func(v interface{}) string {
			return fn(v.(Record))
		},
	}
}

func NewMultiValueIndex(name string, fn IndexMultiValueFn) Index {
	return Index{
		name:    name,
		options: DefaultIndexOptions,
		fn: func(v interface{}) []string {
			return fn(v.(Record))
		},
	}
}

func (idx Index) Name() string {
	return idx.name
}

func (idx Index) Options() collection.IndexOption {
	return idx.options
}

func (idx Index) Func() interface{} {
	return idx.fn
}

func (idx Index) Unique() Index {
	if idx.options&collection.UniqueIndex == 0 {
		idx.options |= collection.UniqueIndex
	}
	return idx
}

func (idx Index) CaseInsensitive() Index {
	if idx.options&collection.CaseInsensitiveIndex == 0 {
		idx.options |= collection.CaseInsensitiveIndex
	}
	return idx
}
