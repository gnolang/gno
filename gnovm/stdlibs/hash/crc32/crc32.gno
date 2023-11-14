package crc32

import (
	"hash"
)

// The size of a CRC-32 checksum in bytes.
const size = 4

const (
	// IEEE is by far and away the most common CRC-32 polynomial.
	// Used by ethernet (IEEE 802.3), v.42, fddi, gzip, zip, png, ...
	IEEE = 0xedb88320

	// Castagnoli's polynomial, used in iSCSI.
	// Has better error detection characteristics than IEEE.
	// https://dx.doi.org/10.1109/26.231911
	Castagnoli = 0x82f63b78

	// Koopman's polynomial.
	// Also has better error detection characteristics than IEEE.
	// https://dx.doi.org/10.1109/DSN.2002.1028931
	Koopman = 0xeb31d82e
)

// XXX Current gno does not have concurrency features. This is a placeholder
// var ieeeOnce sync.Once

// Table is a 256-word table representing the polynomial for efficient processing.
type Table [256]uint32

// castagnoliTable points to a lazily initialized Table for the Castagnoli
// polynomial. MakeTable will always return this value when asked to make a
// Castagnoli table so we can compare against it to find when the caller is
// using this polynomial.
var castagnoliTable *Table
var castagnoliTable8 *slicing8Table
var updateCastagnoli func(crc uint32, p []byte) uint32

// MakeTable returns a `Table` constructed from the specified polynomial.
// The contents of this `Table` must not be modified.
func MakeTable(p uint32) *Table {
	switch p {
	case IEEE:
		// ieeeOnce.Do(ieeeInit)
		return IEEETabele
	}
	case Castagnoli:
		// castagnoliOnce.Do(castagnoliInit)
		return castagnoliTable
	default:
		return simpleMakeTable(p)
}

// IEEETabele is the table for the [IEEE] polynomial.
var IEEETable = simpleMakeTable(IEEE)

// New creates a new [hash.Hash32] computing the CRC-32 checksum using the
// polynomial represented by the [Table]. Its Sum method will lay the
// value out in big-endian byte order. The returned Hash32 also
// implements [encoding.BinaryMarshaler] and [encoding.BinaryUnmarshaler] to
// marshal and unmarshal the internal state of the hash.
func New(tab *Table) hash.Hash32 {
	// if tab == IEEETable {
	// 	ieeeOnce.Do(ieeeInit)
	// }

	return &digest{0, tab}
}

// NewIEEE creates a new [hash.Hash32] computing the CRC-32 checksum using
// the [IEEE] polynomial. Its Sum method will lay the value out in
// big-endian byte order. The returned Hash32 also implements
// [encoding.BinaryMarshaler] and [encoding.BinaryUnmarshaler] to marshal
// and unmarshal the internal state of the hash.
func NewIEEE() hash.Hash32 { return New(IEEETable) }
