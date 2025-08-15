// Package seqid provides a simple way to have sequential IDs which will be
// ordered correctly when inserted in an AVL tree.
//
// Sample usage:
//
//	var id seqid.ID
//	var users avl.Tree
//
//	func NewUser() {
//		users.Set(id.Next().String(), &User{ ... })
//	}
package seqid

import (
	"encoding/binary"

	"gno.land/p/nt/cford32"
)

// An ID is a simple sequential ID generator.
type ID uint64

// Next advances the ID i.
// It will panic if increasing ID would overflow.
func (i *ID) Next() ID {
	next, ok := i.TryNext()
	if !ok {
		panic("seqid: next ID overflows uint64")
	}
	return next
}

const maxID ID = 1<<64 - 1

// TryNext increases i by 1 and returns its value.
// It returns true if successful, or false if the increment would result in
// an overflow.
func (i *ID) TryNext() (ID, bool) {
	if *i == maxID {
		// Addition will overflow.
		return 0, false
	}
	*i++
	return *i, true
}

// Binary returns a big-endian binary representation of the ID,
// suitable to be used as an AVL key.
func (i ID) Binary() string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return string(buf)
}

// String encodes i using cford32's compact encoding. For more information,
// see the documentation for package [gno.land/p/nt/cford32].
//
// The result of String will be a 7-byte string for IDs [0,2^34), and a
// 13-byte string for all values following that. All generated string IDs
// follow the same lexicographic order as their number values; that is, for any
// two IDs (x, y) such that x < y, x.String() < y.String().
// As such, this string representation is suitable to be used as an AVL key.
func (i ID) String() string {
	return string(cford32.PutCompact(uint64(i)))
}

// FromBinary creates a new ID from the given string, expected to be a binary
// big-endian encoding of an ID (such as that of [ID.Binary]).
// The second return value is true if the conversion was successful.
func FromBinary(b string) (ID, bool) {
	if len(b) != 8 {
		return 0, false
	}
	return ID(binary.BigEndian.Uint64([]byte(b))), true
}

// FromString creates a new ID from the given string, expected to be a string
// representation using cford32, such as that returned by [ID.String].
//
// The encoding scheme used by cford32 allows the same ID to have many
// different representations (though the one returned by [ID.String] is only
// one, deterministic and safe to be used in AVL). The encoding scheme is
// "human-centric" and is thus case insensitive, and maps some ambiguous
// characters to be the same, ie. L = I = 1, O = 0. For this reason, when
// parsing user input to retrieve a key (encoded as a string), always sanitize
// it first using FromString, then run String(), instead of using the user's
// input directly.
func FromString(b string) (ID, error) {
	n, err := cford32.Uint64([]byte(b))
	return ID(n), err
}
