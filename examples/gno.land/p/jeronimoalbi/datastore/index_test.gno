package datastore

import (
	"testing"

	"gno.land/p/moul/collection"
	"gno.land/p/nt/uassert"
)

func TestNewIndex(t *testing.T) {
	cases := []struct {
		name    string
		options collection.IndexOption
		setup   func(Index) Index
	}{
		{
			name:    "default",
			options: DefaultIndexOptions,
		},
		{
			name:    "unique",
			options: DefaultIndexOptions | collection.UniqueIndex,
			setup:   func(idx Index) Index { return idx.Unique() },
		},
		{
			name:    "case insensitive",
			options: DefaultIndexOptions | collection.CaseInsensitiveIndex,
			setup:   func(idx Index) Index { return idx.CaseInsensitive() },
		},
		{
			name:    "unique case insensitive",
			options: DefaultIndexOptions | collection.CaseInsensitiveIndex | collection.UniqueIndex,
			setup:   func(idx Index) Index { return idx.CaseInsensitive().Unique() },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name := "foo"
			idx := NewIndex(name, func(Record) string { return "" })

			if tc.setup != nil {
				idx = tc.setup(idx)
			}

			uassert.Equal(t, name, idx.Name())
			uassert.Equal(t, uint64(tc.options), uint64(idx.Options()))

			_, ok := idx.Func().(func(interface{}) string)
			uassert.True(t, ok)
		})
	}
}

func TestNewMultiIndex(t *testing.T) {
	cases := []struct {
		name    string
		options collection.IndexOption
		setup   func(Index) Index
	}{
		{
			name:    "default",
			options: DefaultIndexOptions,
		},
		{
			name:    "unique",
			options: DefaultIndexOptions | collection.UniqueIndex,
			setup:   func(idx Index) Index { return idx.Unique() },
		},
		{
			name:    "case insensitive",
			options: DefaultIndexOptions | collection.CaseInsensitiveIndex,
			setup:   func(idx Index) Index { return idx.CaseInsensitive() },
		},
		{
			name:    "unique case insensitive",
			options: DefaultIndexOptions | collection.CaseInsensitiveIndex | collection.UniqueIndex,
			setup:   func(idx Index) Index { return idx.CaseInsensitive().Unique() },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name := "foo"
			idx := NewMultiValueIndex(name, func(Record) []string { return nil })

			if tc.setup != nil {
				idx = tc.setup(idx)
			}

			uassert.Equal(t, name, idx.Name())
			uassert.Equal(t, uint64(tc.options), uint64(idx.Options()))

			_, ok := idx.Func().(func(interface{}) []string)
			uassert.True(t, ok)
		})
	}
}
