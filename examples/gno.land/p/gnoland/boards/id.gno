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

// IdentifierGenerator defines an interface for sequential unique identifier generators.
type IdentifierGenerator interface {
	// Current returns the last generated ID.
	Last() ID

	// Next generates a new ID or panics if increasing ID overflows.
	Next() ID
}

// NewIdentifierGenerator creates a new sequential unique identifier generator.
func NewIdentifierGenerator() IdentifierGenerator {
	return &idGenerator{}
}

type idGenerator struct {
	last seqid.ID
}

// Current returns the last generated ID.
func (g idGenerator) Last() ID {
	return ID(g.last)
}

// Next generates a new ID or panics if increasing ID overflows.
func (g *idGenerator) Next() ID {
	return ID(g.last.Next())
}
