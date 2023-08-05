package binutils

import (
	"encoding/binary"
	"errors"
)

var ErrInvalidLengthPrefixedString = errors.New("invalid length-prefixed string")

func EncodeLengthPrefixedStringUint16BE(s string) []byte {
	b := make([]byte, 2+len(s))
	binary.BigEndian.PutUint16(b, uint16(len(s)))
	copy(b[2:], s)
	return b
}

func DecodeLengthPrefixedStringUint16BE(b []byte) (string, []byte, error) {
	if len(b) < 2 {
		return "", nil, ErrInvalidLengthPrefixedString
	}
	l := binary.BigEndian.Uint16(b)
	if len(b) < 2+int(l) {
		return "", nil, ErrInvalidLengthPrefixedString
	}
	return string(b[2 : 2+l]), b[l+2:], nil
}

func MustDecodeLengthPrefixedStringUint16BE(b []byte) (string, []byte) {
	s, r, err := DecodeLengthPrefixedStringUint16BE(b)
	if err != nil {
		panic(err)
	}
	return s, r
}
