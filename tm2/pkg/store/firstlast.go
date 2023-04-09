package store

import (
	"bytes"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/types"
)

// Gets the first item.
func First(st Store, start, end []byte) (kv std.KVPair, ok bool) {
	iter, err := st.Iterator(start, end)
	if err != nil {
		panic(err)
	}

	if !iter.Valid() {
		return kv, false
	}
	defer iter.Close()

	return std.KVPair{Key: iter.Key(), Value: iter.Value()}, true
}

// Gets the last item.  `end` is exclusive.
func Last(st Store, start, end []byte) (kv std.KVPair, ok bool) {
	iter, err := st.ReverseIterator(end, start)
	if err != nil {
		panic(err)
	}

	if !iter.Valid() {
		v, err := st.Get(start)
		if err != nil {
			panic(err)
		}

		if v != nil {
			return std.KVPair{Key: types.Cp(start), Value: types.Cp(v)}, true
		}
		return kv, false
	}
	defer iter.Close()

	if bytes.Equal(iter.Key(), end) {
		// Skip this one, end is exclusive.
		iter.Next()
		if !iter.Valid() {
			return kv, false
		}
	}

	return std.KVPair{Key: iter.Key(), Value: iter.Value()}, true
}
