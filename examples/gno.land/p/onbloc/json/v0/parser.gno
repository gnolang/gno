package json

import (
	"bytes"
)

const (
	unescapeStackBufSize = 64
	absMinInt64          = 1 << 63
	maxInt64             = absMinInt64 - 1
	maxUint64            = 1<<64 - 1
)

// PaseStringLiteral parses a string from the given byte slice.
func ParseStringLiteral(data []byte) (string, error) {
	var buf [unescapeStackBufSize]byte

	bf, err := Unescape(data, buf[:])
	if err != nil {
		return "", errInvalidStringInput
	}

	return string(bf), nil
}

// ParseBoolLiteral parses a boolean value from the given byte slice.
func ParseBoolLiteral(data []byte) (bool, error) {
	switch {
	case bytes.Equal(data, trueLiteral):
		return true, nil
	case bytes.Equal(data, falseLiteral):
		return false, nil
	default:
		return false, errMalformedBooleanValue
	}
}
