package std

import (
	"bytes"
	"sort"
)

//----------------------------------------
// KVPair

// KVPair is a key-value struct for []byte value.
type KVPair struct {
	Key   []byte
	Value []byte
}

// KVPairs is a slice of KVPair.
type KVPairs []KVPair

// Len returns the length of kvs.
func (kvs KVPairs) Len() int {
	return len(kvs)
}

// Less reports whether kvs[i] should be ordered before kvs[j].
func (kvs KVPairs) Less(i, j int) bool {
	switch bytes.Compare(kvs[i].Key, kvs[j].Key) {
	case -1:
		return true
	case 0:
		return bytes.Compare(kvs[i].Value, kvs[j].Value) < 0
	case 1:
		return false
	default:
		panic("invalid comparison result")
	}
}

// Swap swaps the elements with indexes, i and j.
func (kvs KVPairs) Swap(i, j int) {
	kvs[i], kvs[j] = kvs[j], kvs[i]
}

// Sort sorts a kvs in ascending order.
func (kvs KVPairs) Sort() {
	sort.Sort(kvs)
}

//----------------------------------------
// KI64Pair

// KI64Pair is a key-value struct for int64 value.
type KI64Pair struct {
	Key   []byte
	Value int64
}

// KI64Pairs is a slice of KI64Pair.
type KI64Pairs []KI64Pair

// Len returns the length of kvs.
func (kvs KI64Pairs) Len() int { return len(kvs) }

// Less reports whether kvs[i] should be ordered before kvs[j].
func (kvs KI64Pairs) Less(i, j int) bool {
	switch bytes.Compare(kvs[i].Key, kvs[j].Key) {
	case -1:
		return true
	case 0:
		return kvs[i].Value < kvs[j].Value
	case 1:
		return false
	default:
		panic("invalid comparison result")
	}
}

// Swap swaps the elements with indexes, i and j.
func (kvs KI64Pairs) Swap(i, j int) {
	kvs[i], kvs[j] = kvs[j], kvs[i]
}

// Sort sorts a kvs in ascending order.
func (kvs KI64Pairs) Sort() {
	sort.Sort(kvs)
}
