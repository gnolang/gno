package amino

import (
	"io"
	"time"
)

// EncodeFieldNumberAndTyp3 writes a protobuf field key (field number + wire type).
// This is the exported wrapper around encodeFieldNumberAndTyp3 for use by
// generated code (genproto2).
func EncodeFieldNumberAndTyp3(w io.Writer, num uint32, typ Typ3) error {
	return encodeFieldNumberAndTyp3(w, num, typ)
}

// TimeSize returns the encoded byte size of a time.Time value
// (the bare content, without field key or length prefix).
func TimeSize(t time.Time) int {
	s := t.Unix()
	ns := int32(t.Nanosecond())
	var n int
	if s != 0 {
		n += UvarintSize(uint64(1)<<3|uint64(Typ3Varint)) + UvarintSize(uint64(s))
	}
	if ns != 0 {
		n += UvarintSize(uint64(2)<<3|uint64(Typ3Varint)) + UvarintSize(uint64(ns))
	}
	return n
}

// DurationSize returns the encoded byte size of a time.Duration value
// (the bare content, without field key or length prefix).
func DurationSize(d time.Duration) int {
	sns := d.Nanoseconds()
	sec, nsec := sns/1e9, int32(sns%1e9)
	var n int
	if sec != 0 {
		n += UvarintSize(uint64(1)<<3|uint64(Typ3Varint)) + UvarintSize(uint64(sec))
	}
	if nsec != 0 {
		n += UvarintSize(uint64(2)<<3|uint64(Typ3Varint)) + UvarintSize(uint64(nsec))
	}
	return n
}
