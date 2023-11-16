// Package seqid provides a simple way to have sequential IDs which will be
// ordered correctly when inserted in an AVL tree.
//
// Sample usage:
//
//	var id seqid.ID
//	var users avl.Tree
//
//	func NewUser() {
//		users.Set(id.Next().Binary(), &User{ ... })
//	}
package seqid

import "encoding/binary"

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

// FromBinary creates a new ID from the given string.
func FromBinary(b string) (ID, bool) {
	if len(b) != 8 {
		return 0, false
	}
	return ID(binary.BigEndian.Uint64([]byte(b))), true
}
