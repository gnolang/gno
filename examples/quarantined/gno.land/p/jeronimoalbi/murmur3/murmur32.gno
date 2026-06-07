// Package murmur3 implements the MurmurHash3 hash algorithm.
//
// Package implements Austin Appleby's MurmurHash3 algorithm.
//
// MurmurHash3 is a fast, non-cryptographic hash function well suited for hash
// tables, bloom filters, data partitioning, and deduplication where speed and
// good distribution matter more than cryptographic security.
//
// References:
//
//	https://en.wikipedia.org/wiki/MurmurHash#Algorithm
//	https://github.com/aappleby/smhasher/blob/master/src/MurmurHash3.cpp
package murmur3

import "strconv"

// EncodeToString returns the hexadecimal encoding of a hash value.
func EncodeToString(sum uint64) string {
	return strconv.FormatUint(sum, 16)
}
