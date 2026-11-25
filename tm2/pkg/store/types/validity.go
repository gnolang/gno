package types

import "fmt"

// MaxKeyLen is the longest key a cache store accepts. It sits far below
// bptree.MaxKeyLen so a full leaf of bptree.B keys holds at most 32 KiB of keys.
const MaxKeyLen = 1 << 10

// Check if the key is valid(key is not nil, and not over MaxKeyLen)
func AssertValidKey(key []byte) {
	if key == nil {
		panic("key is nil")
	}
	// Refused while the tx runs: the tree refuses a long key only when the
	// commit flushes the caches, and a panic there halts the chain.
	if len(key) > MaxKeyLen {
		panic(fmt.Sprintf("store key length %d exceeds the maximum %d", len(key), MaxKeyLen))
	}
}

// Check if the value is valid(value is not nil)
func AssertValidValue(value []byte) {
	if value == nil {
		panic("value is nil")
	}
}
