package types

import (
	"bytes"
)

// Iterator over all the keys with a certain prefix in ascending order
func PrefixIterator(kvs Store, prefix []byte) Iterator {
	return kvs.Iterator(prefix, PrefixEndBytes(prefix))
}

// Iterator over all the keys with a certain prefix in descending order.
func ReversePrefixIterator(kvs Store, prefix []byte) Iterator {
	return kvs.ReverseIterator(prefix, PrefixEndBytes(prefix))
}

// Compare two stores, return either the first key/value pair
// at which they differ and whether or not they are equal, skipping
// value comparison for a set of provided prefixes
func DiffStores(a Store, b Store, prefixesToSkip [][]byte) (kvA KVPair, kvB KVPair, count int64, equal bool) {
	iterA := a.Iterator(nil, nil)
	iterB := b.Iterator(nil, nil)
	count = int64(0)
	for iterA.Valid() || iterB.Valid() {
		var kvA, kvB KVPair
		if iterA.Valid() {
			kvA = KVPair{Key: iterA.Key(), Value: iterA.Value()}
			iterA.Next()
		}
		if iterB.Valid() {
			kvB = KVPair{Key: iterB.Key(), Value: iterB.Value()}
			iterB.Next()
		}
		if !bytes.Equal(kvA.Key, kvB.Key) {
			return kvA, kvB, count, false
		}
		compareValue := true
		for _, prefix := range prefixesToSkip {
			// Skip value comparison if we matched a prefix
			if bytes.Equal(kvA.Key[:len(prefix)], prefix) {
				compareValue = false
			}
		}
		if compareValue && !bytes.Equal(kvA.Value, kvB.Value) {
			return kvA, kvB, count, false
		}
		count++
	}
	return KVPair{}, KVPair{}, count, true
}

// PrefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func PrefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for {
		if end[len(end)-1] != byte(255) {
			end[len(end)-1]++
			break
		} else {
			end = end[:len(end)-1]
			if len(end) == 0 {
				end = nil
				break
			}
		}
	}
	return end
}

// InclusiveEndBytes returns the []byte that would end a
// range query such that the input would be included
func InclusiveEndBytes(inclusiveBytes []byte) (exclusiveBytes []byte) {
	exclusiveBytes = append(inclusiveBytes, byte(0x00))
	return exclusiveBytes
}

// ----------------------------------------
func Cp(bz []byte) (ret []byte) {
	if bz == nil {
		return nil
	}
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}
