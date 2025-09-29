package boards

import (
	"strconv"
	"strings"

	"gno.land/p/nt/seqid"
)

const paddedStringLen = 10

// ID defines a type for unique identifiers.
type ID uint64

// String returns the ID as a string.
func (id ID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

// PaddedString returns the ID as a 10 character string padded with zeroes.
// This value can be used for indexing by ID.
func (id ID) PaddedString() string {
	s := id.String()
	return strings.Repeat("0", paddedStringLen-len(s)) + s
}

// Key returns the ID as a string which can be used to index by ID.
func (id ID) Key() string {
	return seqid.ID(id).String()
}
