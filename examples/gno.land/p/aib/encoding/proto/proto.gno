package proto

import (
	"encoding/binary"
	"time"

	"gno.land/p/nt/ufmt"
)

type WireType int

const (
	VARINT  WireType = 0 //	int32, int64, uint32, uint64, sint32, sint64, bool, enum
	FIXED64 WireType = 1 // fixed64, sfixed64, double
	LEN     WireType = 2 // string, bytes, embedded messages, packed repeated fields
	FIXED32 WireType = 5 //	fixed32, sfixed32, float
)

// AppendVarint appends a varint field.
func AppendVarint(buf []byte, fieldNum int, v uint64) []byte {
	if v == 0 {
		return buf
	}
	buf = AppendTag(buf, fieldNum, VARINT)
	return binary.AppendUvarint(buf, v)
}

// AppendFixed64 appends a fixed 64-bit field.
func AppendFixed64(buf []byte, fieldNum int, v uint64) []byte {
	if v == 0 {
		return buf
	}
	buf = AppendTag(buf, fieldNum, FIXED64)
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	return append(buf, b[:]...)
}

// AppendLengthDelimited appends a length-delimited field.
func AppendLengthDelimited(buf []byte, fieldNum int, bz []byte) []byte {
	if len(bz) == 0 {
		return buf
	}
	buf = AppendTag(buf, fieldNum, LEN)
	buf = binary.AppendUvarint(buf, uint64(len(bz)))
	return append(buf, bz...)
}

// AppendTime appends a google.protobuf.Timestamp field.
func AppendTime(buf []byte, fieldNum int, t time.Time) []byte {
	var (
		_buf    []byte
		seconds = t.Unix()
		nanos   = int32(t.Nanosecond())
	)
	// Field 1: seconds (int64 - varint)
	_buf = AppendVarint(_buf, 1, uint64(seconds))
	// Field 2: nanos (int32 - varint)
	_buf = AppendVarint(_buf, 2, uint64(nanos))

	return AppendLengthDelimited(buf, fieldNum, _buf)
}

// AppendTag appends a protobuf tag (field number and wire type)
func AppendTag(buf []byte, fieldNum int, wireType WireType) []byte {
	tag := (fieldNum << 3) | int(wireType)
	return binary.AppendUvarint(buf, uint64(tag))
}

// DecodeVarint reads a varint from the byte slice and returns the value and new position
func DecodeVarint(buf []byte, pos int) (uint64, int, error) {
	var result uint64
	var shift uint
	for {
		if pos >= len(buf) {
			return 0, pos, ufmt.Errorf("buffer underflow while reading varint")
		}
		b := buf[pos]
		pos++
		result |= uint64(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return result, pos, nil
}

// DecodeString reads a string (length-prefixed) from the byte slice
func DecodeString(buf []byte, pos int) (string, int, error) {
	length, newPos, err := DecodeVarint(buf, pos)
	if err != nil {
		return "", newPos, err
	}
	pos = newPos

	if pos+int(length) > len(buf) {
		return "", pos, ufmt.Errorf("buffer underflow while reading string")
	}

	str := string(buf[pos : pos+int(length)])
	return str, pos + int(length), nil
}
