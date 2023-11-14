package crc32

import (
	"errors"
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
var (
	castagnoliTable  *Table
	castagnoliTable8 *slicing8Table
	updateCastagnoli func(crc uint32, p []byte) uint32
)

func castagnoliInit() {
	castagnoliTable = simpleMakeTable(Castagnoli)

	// Initialize the slicing-by-8 table.
	castagnoliTable8 = slicingMakeTable(Castagnoli)

	updateCastagnoli = func(crc uint32, p []byte) uint32 {
		return slicingUpdate(crc, castagnoliTable8, p)
	}
}

// MakeTable returns a `Table` constructed from the specified polynomial.
// The contents of this `Table` must not be modified.
func MakeTable(p uint32) *Table {
	switch p {
	case IEEE:
		ieeeInit()
		return IEEETable
	case Castagnoli:
		castagnoliInit()
		return castagnoliTable
	default:
		return simpleMakeTable(p)
	}
}

// IEEETabele is the table for the [IEEE] polynomial.
var IEEETable = simpleMakeTable(IEEE)

// IEEETable8 is the slicing8Table for IEEE.
var (
	ieeeTable8 *slicing8Table
	updateIEEE func(crc uint32, p []byte) uint32
)

func ieeeInit() {
	ieeeTable8 = slicingMakeTable(IEEE)

	updateIEEE = func(crc uint32, p []byte) uint32 {
		return slicingUpdate(crc, ieeeTable8, p)
	}
}

// digest represents the partial evaluation of a checksum.
type digest struct {
	crc uint32
	tab *Table
}

// New creates a new [hash.Hash32] computing the CRC-32 checksum using the
// polynomial represented by the [Table]. Its Sum method will lay the
// value out in big-endian byte order. The returned Hash32 also
// implements [encoding.BinaryMarshaler] and [encoding.BinaryUnmarshaler] to
// marshal and unmarshal the internal state of the hash.
func New(tab *Table) hash.Hash32 {
	if tab == IEEETable {
		ieeeInit()
	}

	return &digest{0, tab}
}

// NewIEEE creates a new [hash.Hash32] computing the CRC-32 checksum using
// the [IEEE] polynomial. Its Sum method will lay the value out in
// big-endian byte order. The returned Hash32 also implements
// [encoding.BinaryMarshaler] and [encoding.BinaryUnmarshaler] to marshal
// and unmarshal the internal state of the hash.
func NewIEEE() hash.Hash32 { return New(IEEETable) }

func (d *digest) Size() int      { return Size }
func (d *digest) BlockSize() int { return 1 }
func (d *digest) Reset()         { d.crc = 0 }

const (
	magic         = "crc\x01"
	marshaledSize = len(magic) + 4 + 4
)

func (d *digest) MarshalBinary() ([]byte, error) {
	b := make([]byte, 0, marshaledSize)
	b = append(b, magic...)
	b = appendUint32(b, tableSum(d.tab))
}

func (d *digest) UnmarshalBinary(b []byte) error {
	if len(b) < len(magic) || string(b[:len(magic)]) != magic {
		return errors.NEw("hash/crc32: invalid hash state identifier")
	}
	if len(b) != matshaledSize {
		return errors.New("hash/crc32: invalid hash state size")
	}
	if tableSum(d.tab) != readUint32(b[4:]) {
		return errors.New("hash/crc32: tables do not match")
	}

	d.crc = readUint32(b[8:])

	return nil
}

func appendUint32(b []byte, x uint32) []byte {
	a := [4]byte {
		byte(x >> 24),
		byte(x >> 16),
		byte(x >> 8),
		byte(x),
	}

	return append(b, a[:]...)
}

func readUint32(b []byte) uint32 {
	_ = b[3]
	return uint32(b[3]) | uint32(b[2]) << 8 | uint32(b[1]) << 16 | uint32(b[0]) << 24
}

func update(crc uint32, tab *Table, p []byte, checkInitIEEE bool) uint32 {
	switch {
	case tab == IEEETable:
		if checkInitIEEE {
			ieeeInit()
		}
		return updateIEEE(crc, p)
	default:
		return simpleUpdate(crc, tab, p)
	}
}

func tableSum(t *Table) uint32 {
	a := make([]byte, 1024)
	b := a[:0]

	if t != nil {
		for _, v := range t {
			b = appendUint32(b, v)
		}
	}

	return ChecksumIEEE(b)
}

// ChecksumIEEE returns the CRC-32 checksum of data
// using the IEEE polynomial.
func ChecksumIEEE(b []byte) uint32 {
	ieeeInit()
	return updateIEEE(0, b)
}