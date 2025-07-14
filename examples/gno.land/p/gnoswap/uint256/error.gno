package uint256

import (
	"errors"
)

var (
	ErrEmptyString      = errors.New("empty hex string")
	ErrSyntax           = errors.New("invalid hex string")
	ErrRange            = errors.New("number out of range")
	ErrMissingPrefix    = errors.New("hex string without 0x prefix")
	ErrEmptyNumber      = errors.New("hex string \"0x\"")
	ErrLeadingZero      = errors.New("hex number with leading zero digits")
	ErrBig256Range      = errors.New("hex number > 256 bits")
	ErrBadBufferLength  = errors.New("bad ssz buffer length")
	ErrBadEncodedLength = errors.New("bad ssz encoded length")
	ErrInvalidBase      = errors.New("invalid base")
	ErrInvalidBitSize   = errors.New("invalid bit size")
)

type u256Error struct {
	fn    string // function name
	input string
	err   error
}

func (e *u256Error) Error() string {
	return e.fn + ": " + e.input + ": " + e.err.Error()
}

func (e *u256Error) Unwrap() error {
	return e.err
}

func errEmptyString(fn, input string) error {
	return &u256Error{fn: fn, input: input, err: ErrEmptyString}
}

func errSyntax(fn, input string) error {
	return &u256Error{fn: fn, input: input, err: ErrSyntax}
}

func errMissingPrefix(fn, input string) error {
	return &u256Error{fn: fn, input: input, err: ErrMissingPrefix}
}

func errEmptyNumber(fn, input string) error {
	return &u256Error{fn: fn, input: input, err: ErrEmptyNumber}
}

func errLeadingZero(fn, input string) error {
	return &u256Error{fn: fn, input: input, err: ErrLeadingZero}
}

func errRange(fn, input string) error {
	return &u256Error{fn: fn, input: input, err: ErrRange}
}

func errBig256Range(fn, input string) error {
	return &u256Error{fn: fn, input: input, err: ErrBig256Range}
}

func errBadBufferLength(fn, input string) error {
	return &u256Error{fn: fn, input: input, err: ErrBadBufferLength}
}

func errInvalidBase(fn string, base int) error {
	return &u256Error{fn: fn, input: string(base), err: ErrInvalidBase}
}

func errInvalidBitSize(fn string, bitSize int) error {
	return &u256Error{fn: fn, input: string(bitSize), err: ErrInvalidBitSize}
}
